package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

// ContainerConfig holds the configuration for running a container
type ContainerConfig struct {
	Image   string            // Docker image name (e.g., "soverstack/runtime:v1.0.0")
	Args    []string          // Command arguments to pass to the container
	EnvVars []string          // Environment variables in "KEY=VALUE" format
	WorkDir string            // Host working directory to mount
}

// RunContainer creates and runs a Docker container with full I/O forwarding.
// This is the core of the launcher - it must:
// 1. Create container with proper configuration
// 2. Attach I/O BEFORE starting (critical for capturing all output)
// 3. Stream stdin/stdout/stderr bidirectionally
// 4. Handle signals (Ctrl+C) gracefully
// 5. Exit with the container's exit code
//
// The container is configured with AutoRemove=true, so it will be automatically
// cleaned up when it exits.
func RunContainer(ctx context.Context, config ContainerConfig) error {
	cli, err := NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	// Get current working directory for volume mount
	cwd := config.WorkDir
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Create container configuration
	containerConfig := &container.Config{
		Image:        config.Image,
		Cmd:          config.Args,
		Env:          config.EnvVars,
		Tty:          false, // Set to false to properly capture output
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		StdinOnce:    false,
		WorkingDir:   "/workspace",
	}

	// Create host configuration (volumes, auto-remove, user mapping)
	hostConfig := &container.HostConfig{
		AutoRemove: true, // Automatically remove container on exit
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: cwd,
				Target: "/workspace",
			},
		},
	}

	// On Linux, run container with same UID/GID as host user to avoid permission issues
	// Windows doesn't need this as Docker Desktop handles permissions automatically
	if runtime.GOOS == "linux" {
		hostConfig.UsernsMode = "host"
		// Note: We could set specific UID:GID here, but "host" mode is simpler
		// and works well for most cases. For production, consider:
		// hostConfig.User = fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	}

	// Create container
	resp, err := cli.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil, // network config
		nil, // platform
		"",  // container name (let Docker generate)
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID

	// Setup signal handling for graceful shutdown (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, stopping container...")

		// Give container 10 seconds to stop gracefully
		timeout := 10
		stopCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		cli.ContainerStop(stopCtx, containerID, container.StopOptions{
			Timeout: &timeout,
		})
	}()

	// CRITICAL: Attach to container I/O BEFORE starting it
	// If we attach after start, we lose initial output
	attachResp, err := cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   true, // Get logs from container start
	})
	if err != nil {
		return fmt.Errorf("failed to attach to container: %w", err)
	}
	defer attachResp.Close()

	// Start the container
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Stream I/O bidirectionally
	// Start goroutines to copy streams concurrently
	errChan := make(chan error, 2)

	// Container stdout/stderr → Host stdout/stderr
	go func() {
		// Docker multiplexes stdout and stderr into a single stream
		// We need to demultiplex it
		_, err := io.Copy(os.Stdout, attachResp.Reader)
		if err != nil && err != io.EOF {
			errChan <- fmt.Errorf("error copying container output: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Host stdin → Container stdin
	go func() {
		_, err := io.Copy(attachResp.Conn, os.Stdin)
		if err != nil && err != io.EOF {
			errChan <- fmt.Errorf("error copying input to container: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for container to finish
	statusCh, waitErrCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-waitErrCh:
		if err != nil {
			return fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		// Container finished, exit with its exit code
		if status.StatusCode != 0 {
			// Non-zero exit code - the CLI had an error
			// We exit with the same code to propagate the error
			os.Exit(int(status.StatusCode))
		}
	}

	// Check for I/O copy errors
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			// Don't fail on I/O errors if container succeeded
			// Just log them
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}

	return nil
}
