package docker

import (
	"context"
	"fmt"
	"runtime"

	"github.com/docker/docker/client"
)

// NewClient creates a new Docker client using environment configuration.
// This uses the standard Docker environment variables:
// - DOCKER_HOST (default: unix:///var/run/docker.sock or npipe:////./pipe/docker_engine on Windows)
// - DOCKER_API_VERSION
// - DOCKER_CERT_PATH
// - DOCKER_TLS_VERIFY
func NewClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return cli, nil
}

// CheckAvailable verifies that Docker daemon is running and accessible.
// Returns an actionable error message if Docker is not available.
func CheckAvailable(ctx context.Context) error {
	cli, err := NewClient()
	if err != nil {
		return formatDockerError(err)
	}
	defer cli.Close()

	// Ping Docker daemon
	_, err = cli.Ping(ctx)
	if err != nil {
		return formatDockerError(err)
	}

	return nil
}

// formatDockerError provides platform-specific actionable error messages
func formatDockerError(err error) error {
	baseMsg := fmt.Sprintf("Docker is not available: %v\n\n", err)

	switch runtime.GOOS {
	case "windows":
		return fmt.Errorf("%sSuggestion: Start Docker Desktop for Windows", baseMsg)
	case "darwin":
		return fmt.Errorf("%sSuggestion: Start Docker Desktop for macOS", baseMsg)
	case "linux":
		return fmt.Errorf("%sSuggestion:\n  - Check if Docker daemon is running: sudo systemctl status docker\n  - Start Docker: sudo systemctl start docker\n  - Ensure your user is in the 'docker' group: sudo usermod -aG docker $USER", baseMsg)
	default:
		return fmt.Errorf("%sSuggestion: Ensure Docker is installed and running", baseMsg)
	}
}
