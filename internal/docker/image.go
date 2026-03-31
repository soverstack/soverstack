package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/image"
)

// PullImage pulls a Docker image from the registry with progress display.
// If the image already exists locally, Docker will use the cached version.
//
// imageName should be in the format: "repository:tag" (e.g., "soverstack/runtime:v1.0.0")
func PullImage(ctx context.Context, imageName string) error {
	cli, err := NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	fmt.Printf("Pulling image: %s\n", imageName)

	// Pull image with progress output
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w\nSuggestion: Check your internet connection and verify the version in platform.yaml", imageName, err)
	}
	defer reader.Close()

	// Parse and display progress
	if err := displayPullProgress(reader); err != nil {
		return err
	}

	fmt.Println("\nImage pulled successfully")
	return nil
}

// displayPullProgress reads the JSON stream from Docker and displays progress information
func displayPullProgress(reader io.ReadCloser) error {
	decoder := json.NewDecoder(reader)

	type ProgressDetail struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	}

	type ProgressMessage struct {
		Status         string         `json:"status"`
		Progress       string         `json:"progress"`
		ProgressDetail ProgressDetail `json:"progressDetail"`
		ID             string         `json:"id"`
		Error          string         `json:"error"`
	}

	for {
		var msg ProgressMessage
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading pull progress: %w", err)
		}

		// Check for errors in the progress stream
		if msg.Error != "" {
			return fmt.Errorf("image pull error: %s", msg.Error)
		}

		// Display progress (simple version - just show status updates)
		if msg.Status != "" {
			if msg.Progress != "" {
				// Display layer progress with carriage return to update same line
				fmt.Fprintf(os.Stderr, "\r%s %s: %s", msg.ID, msg.Status, msg.Progress)
			} else if msg.ID != "" {
				// Display layer status
				fmt.Fprintf(os.Stderr, "\r%s: %s", msg.ID, msg.Status)
			} else {
				// Display general status
				fmt.Fprintf(os.Stderr, "\r%s", msg.Status)
			}
		}
	}

	return nil
}
