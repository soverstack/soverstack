package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/soverstack/launcher/internal/docker"
	"github.com/soverstack/launcher/internal/platform"
)

const (
	// Docker image repository (GitHub Container Registry)
	imageRepository = "ghcr.io/soverstack/cli-runtime"

	// Default timeout for Docker operations
	defaultTimeout = 5 * time.Minute
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Create context with timeout for Docker operations
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Step 1: Capture all arguments (everything after the binary name)
	// Examples:
	//   soverstack validate platform.yaml → args = ["validate", "platform.yaml"]
	//   soverstack plan platform.yaml → args = ["plan", "platform.yaml"]
	//   soverstack dns:update example.com → args = ["dns:update", "example.com"]
	args := os.Args[1:]

	// Handle version flag
	if len(args) > 0 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("soverstack launcher version %s\n", Version)
		return nil
	}

	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printHelp()
		return nil
	}

	// Step 2: Extract version from platform.yaml
	// Falls back to launcher's own version + "-SNAPSHOT" if not found
	version, err := platform.ExtractVersion(args)
	if err != nil || version == "latest" {
		version = Version + "-SNAPSHOT"
	}

	// Build full image name
	imageName := fmt.Sprintf("%s:%s", imageRepository, version)

	// Step 3: Check Docker is available
	if err := docker.CheckAvailable(ctx); err != nil {
		return err
	}

	// Step 4: Pull Docker image (or use cached version)
	// Docker automatically caches images, so this is fast on subsequent runs
	if err := docker.PullImage(ctx, imageName); err != nil {
		return err
	}

	// Step 5: Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Step 6: Run container with full I/O forwarding
	// The project directory is mounted at /workspace — the runtime reads
	// .env, platform.yaml, inventory/, etc. directly from the mount.
	// No env var forwarding needed.
	config := docker.ContainerConfig{
		Image:   imageName,
		Args:    args,
		WorkDir: cwd,
	}

	return docker.RunContainer(ctx, config)
}

func printHelp() {
	fmt.Println(`Soverstack Launcher - Native proxy for Soverstack CLI

USAGE:
  soverstack <command> [arguments...]

COMMANDS:
  validate <platform-yaml>     Validate platform configuration
  plan <platform-yaml>         Generate execution plan
  apply                        Apply infrastructure changes
  dns:update <domain>          Update DNS nameservers
  destroy <resource> <id>      Destroy infrastructure resource

FLAGS:
  --version, -v    Show launcher version
  --help, -h       Show this help message

EXAMPLES:
  soverstack validate platform.yaml
  soverstack plan platform.yaml
  soverstack dns:update example.com

HOW IT WORKS:
  The launcher is a transparent proxy that:
  1. Reads platform.yaml to determine runtime version
  2. Pulls the soverstack/runtime:<version> Docker image
  3. Mounts your project directory at /workspace in the container
  4. Runs the CLI inside the container with your arguments

  The runtime reads .env, platform.yaml, inventory/, workloads/,
  and .ssh/ directly from the mounted directory. No environment
  variables are forwarded from the host.

REQUIREMENTS:
  - Docker must be installed and running
  - Internet connection (for first-time image pull)

For more information, visit: https://soverstack.com`)
}
