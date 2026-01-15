package main

import (
	"context"
	config "drive-uploader/config"
	"drive-uploader/googledrive"
	"log/slog"
	"os"
)

const (
	tokenFile    = "google_token.json"
	fileToUpload = "send_file.txt"
	folderID     = "1bdlpF5xWqyNg0vXBLxH5ZbpzDwIkIuw3"
)

func main() {
	ctx := context.Background()

	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	token, err := googledrive.GetToken(config.Oauth2Config, tokenFile)
	if err != nil {
		slog.Error("Ошибка получения токена", "error", err)
		os.Exit(1)
	}

	if err := googledrive.SaveToken(tokenFile, token); err != nil {
		slog.Error("Ошибка сохранения токена", "error", err)
		os.Exit(1)
	}

	driveService, err := googledrive.NewDriveService(ctx, config.Oauth2Config, token)
	if err != nil {
		slog.Error("Ошибка создания сервиса Drive", "error", err)
		os.Exit(1)
	}

	if err := driveService.UploadFile(fileToUpload, folderID); err != nil {
		slog.Error("Ошибка загрузки файла", "error", err)
		os.Exit(1)
	}

	slog.Info("Файл успешно загружен в Google Drive", "file", fileToUpload)
}
