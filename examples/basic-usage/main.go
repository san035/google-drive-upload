package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"drive-uploader/pkg/config"
	"drive-uploader/pkg/googledrive"
)

func main() {
	ctx := context.Background()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем сервис Google Drive
	driveService, err := googledrive.NewDriveService(ctx, cfg.ConfigGoogleDrives)
	if err != nil {
		log.Fatalf("Ошибка создания сервиса Drive: %v", err)
	}

	// Пример 1: Загрузка одного файла
	filename := "example.txt"
	fmt.Printf("Загружаем файл: %s\n", filename)
	if err := driveService.UploadFile(ctx, filename); err != nil {
		log.Fatalf("Ошибка загрузки: %v", err)
	}
	fmt.Printf("Файл %s успешно загружен!\n", filename)

}
