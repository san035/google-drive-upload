package main

import (
	"context"
	"drive-uploader/pkg/config"
	"drive-uploader/pkg/googledrive"
	"log/slog"
	"os"
)

const (
	fileToUpload = "send_file.txt"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	driveService, err := googledrive.NewDriveService(ctx, cfg.ConfigGoogleDrives)
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
