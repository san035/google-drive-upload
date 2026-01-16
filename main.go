package main

import (
	"context"
	"github.com/san035/google-drive-upload/pkg/configgoogledrive"
	"github.com/san035/google-drive-upload/pkg/googleupload"

	"log/slog"
	"os"
)

const (
	fileToUpload = "send_file.txt"
)

func main() {
	ctx := context.Background()

	cfg, err := configgoogledrive.LoadConfig()
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	driveService, err := googleupload.NewDriveService(ctx, cfg.ConfigGoogleDrives)
	if err != nil {
		slog.Error("Ошибка создания сервиса Drive", "error", err)
		os.Exit(1)
	}

	if err := driveService.UploadFile(ctx, fileToUpload); err != nil {
		slog.Error("Ошибка загрузки файла", "error", err)
		os.Exit(1)
	}

	slog.Info("Файл успешно загружен в Google Drive", "file", fileToUpload)
}
