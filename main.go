package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/soverstack/cli-launcher/internal/docker"
	"github.com/soverstack/cli-launcher/internal/selfupdate"
	"github.com/soverstack/cli-launcher/internal/update"
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

	// Check for updates in background (non-blocking, fails silently)
	updateMsg := make(chan string, 1)
	go func() {
		updateMsg <- update.CheckForUpdate(Version)
	}()
	defer func() {
		if msg := <-updateMsg; msg != "" {
			fmt.Fprint(os.Stderr, msg)
		}
	}()

	args := os.Args[1:]

	// Handle version flag
	if len(args) > 0 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("soverstack version %s\n", Version)
		return nil
	}

	// Handle help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printHelp()
		return nil
	}

	// Handle update command (launcher-only, not forwarded to Docker)
	if len(args) > 0 && args[0] == "update" {
		method := selfupdate.Detect()
		fmt.Printf("Current version: %s (installed via %s)\n", Version, method)
		return selfupdate.Run(method)
	}

	// Use the launcher's own version to determine the runtime image
	imageName := fmt.Sprintf("%s:%s", imageRepository, Version)

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
	fmt.Println(`Soverstack - Sovereign infrastructure orchestration

USAGE:
  soverstack <command> [arguments...] [options]

COMMANDS:
  init [project-name]              Initialize a new project
  validate [path]                  Validate project configuration
  plan [path]                      Show execution plan
  apply [path]                     Apply infrastructure changes
  add region [name]                Add a new region
  add zone [region] [zone-name]    Add a new zone to a region
  generate ssh                     Generate or rotate SSH keys
  update                           Update soverstack to the latest version

OPTIONS:
  -v, --verbose    Show detailed output
  --debug          Show debug information
  --version        Show version
  -h, --help       Show this help message

EXAMPLES:
  soverstack init my-infra
  soverstack validate
  soverstack plan --verbose
  soverstack apply
  soverstack add region us --zones portland,seattle
  soverstack generate ssh --all

REQUIREMENTS:
  Docker must be installed and running.

https://soverstack.io`)
}
