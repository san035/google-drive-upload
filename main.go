package main

import (
	"context"

	"github.com/san035/google-drive-upload/pkg/googleupload"

	"log/slog"
	"os"
)

const (
	fileUploadFefault = "send_file.txt"
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

	var fileUpload string
	if len(os.Args) > 1 {
		fileUpload = os.Args[1]
	} else {
		fileUpload = fileUploadFefault
	}

	if err := driveService.UploadFile(ctx, fileUpload); err != nil {
		slog.Error("Ошибка загрузки файла", "error", err)
		os.Exit(1)
	}
}
