package googledrive

import (
	"context"
	"drive-uploader/config"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GoogleDisks struct {
	ListGoogleDisk    []*GoogleDisk
	GoogleDiskDefault *GoogleDisk
}

type GoogleDisk struct {
	Srv *drive.Service
	cfg *config.ConfigGoogleDrive
}

// NewDriveService создаёт новый сервис Drive API
func NewDriveService(ctx context.Context, config config.ConfigGoogleDrives) (*GoogleDisks, error) {
	var (
		gdDefault      *GoogleDisk
		listGoogleDisk = make([]*GoogleDisk, 0, len(config))
	)
	for _, cfg := range config {
		data, err := os.ReadFile(cfg.GoogleCredentialsFile)
		if err != nil {
			return nil, err
		}

		oauth2Config, err := google.ConfigFromJSON(data, drive.DriveScope)
		if err != nil {
			return nil, err
		}

		tokenFile := strings.TrimSuffix(cfg.GoogleCredentialsFile, filepath.Ext(cfg.GoogleCredentialsFile)) + "_token.json"
		token, err := GetToken(oauth2Config, tokenFile)
		if err != nil {
			return nil, err
		}

		client := oauth2Config.Client(ctx, token)
		srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, err
		}

		gd := &GoogleDisk{
			Srv: srv,
			cfg: cfg,
		}
		if gdDefault == nil {
			gdDefault = gd
		}

		listGoogleDisk = append(listGoogleDisk, gd)
	}

	gds := GoogleDisks{
		GoogleDiskDefault: gdDefault,
		ListGoogleDisk:    listGoogleDisk,
	}
	return &gds, nil
}

// UploadFile загружает файл на Google Drive
func (gd *GoogleDisks) UploadFile(ctx context.Context, filename string) error {
	// Получаем информацию о файле
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("ошибка получения информации о файле: %w", err)
	}

	fileSize := fileInfo.Size()

	// Проверяем наличие свободного места
	hasSpace, quota, err := gd.GoogleDiskDefault.HasEnoughSpace(ctx, fileSize)
	if err != nil {
		return fmt.Errorf("ошибка проверки свободного места: %w", err)
	}

	if !hasSpace {
		return fmt.Errorf("недостаточно свободного места на Google Drive. Требуется: %s, свободно: %s (всего: %s, используется: %s)",
			FormatBytes(fileSize), FormatBytes(quota.FreeBytes), FormatBytes(quota.TotalBytes), FormatBytes(quota.UsedBytes))
	}

	// Открываем файл для загрузки
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("ошибка закрытия файла", "filename", filename, "error", err)
		}
	}()

	driveFile := &drive.File{
		Name:    filename,
		Parents: []string{gd.GoogleDiskDefault.cfg.FolderID},
	}

	_, err = gd.GoogleDiskDefault.Srv.Files.Create(driveFile).Media(file).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("ошибка загрузки файла: %w", err)
	}

	slog.Info("Файл успешно загружен",
		"filename", filename,
		"fileSize", FormatBytes(fileSize),
		"freeSpaceAfter", FormatBytes(quota.FreeBytes-fileSize))

	return nil
}
