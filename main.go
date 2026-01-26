package main

import (
	"context"
	"strings"

	"github.com/san035/google-drive-upload/pkg/googleupload"

	"log/slog"
	"os"
)

const (
	fileUploadDefault = "send_file.txt"
	idDiskDefault     = ""
)

func main() {
	ctx := context.Background()

	cfg, err := googleupload.LoadConfig()
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	driveService, err := googleupload.NewDriveService(ctx, cfg)
	if err != nil {
		slog.Error("Ошибка создания сервиса Drive", "error", err)
		os.Exit(1)
	}

	file, diskID := getFileAndIDDisk()

	if err := driveService.UploadFile(ctx, file, diskID); err != nil {
		slog.Error("Ошибка загрузки файла", "error", err)
		os.Exit(1)
	}
}

func getFileAndIDDisk() (file string, diskID string) {
	for _, arg := range os.Args[1:] {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch strings.ToLower(parts[0]) {
		case "file":
			file = parts[1]
		case "iddisk":
			diskID = parts[1]
		}
	}

	if file == "" {
		file = fileUploadDefault
	}
	if diskID == "" {
		diskID = idDiskDefault
	}
	return file, diskID
}
