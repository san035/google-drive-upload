package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/san035/google-drive-upload/pkg/googleupload"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := googleupload.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	// Create Google Drive service
	driveService, err := googleupload.NewDriveService(ctx, cfg)
	if err != nil {
		slog.Error("Failed to create Drive service", slog.Any("error", err))
		os.Exit(1)
	}

	// Example 1: Upload a single file
	filename := "example.txt"
	if err := driveService.UploadFile(ctx, filename); err != nil {
		slog.Error("Failed to upload file", slog.Any("error", err))
		os.Exit(1)
	}
}
