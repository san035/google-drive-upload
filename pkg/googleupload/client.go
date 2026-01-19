package googleupload

import (
	"context"
	"fmt"

	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type GoogleDisks struct {
	ListGoogleDisk    []*GoogleDisk
	GoogleDiskDefault *GoogleDisk
}

type GoogleDisk struct {
	Srv *drive.Service
	cfg *ConfigGoogleDrive
}

type Option struct {
	Id string
}

// NewDriveService создаёт новый сервис Drive API
func NewDriveService(ctx context.Context, config *Config) (*GoogleDisks, error) {
	var (
		gdDefault      *GoogleDisk
		listGoogleDisk = make([]*GoogleDisk, 0, len(config.ConfigGoogleDrives))
	)

	// Получаем хост и порт для OAuth callback
	callbackHostPort := config.OAuthCallbackHostPort

	for _, cfg := range config.ConfigGoogleDrives {
		data, err := os.ReadFile(cfg.GoogleCredentialsFile)
		if err != nil {
			return nil, err
		}

		oauth2Config, err := google.ConfigFromJSON(data, drive.DriveScope)
		if err != nil {
			return nil, err
		}

		// Устанавливаем redirect URL для локального сервера авторизации
		oauth2Config.RedirectURL = "http://" + callbackHostPort + "/oauth2/callback"

		gd := &GoogleDisk{
			cfg: cfg,
		}

		token, err := gd.GetToken(oauth2Config)
		if err != nil {
			return nil, err
		}

		client := oauth2Config.Client(ctx, token)
		gd.Srv, err = drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, err
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
func (gds *GoogleDisks) UploadFile(ctx context.Context, filename string) error {
	// Получаем информацию о файле
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("ошибка получения информации о файле: %w", err)
	}
	fileSize := fileInfo.Size()

	// Удаляем самые старые копии, оставляя UploadCopiesCount - 1 копий
	if err := gds.deleteOldCopies(ctx, filename); err != nil {
		slog.Warn("ошибка удаления старых копий", "filename", filename, "error", err)
		// Не прерываем процесс загрузки, если не удалось удалить старые копии
	}

	// Проверяем наличие свободного места
	hasSpace, quota, err := gds.GoogleDiskDefault.HasEnoughSpace(ctx, fileSize)
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
		Name:    filepath.Base(filename),
		Parents: []string{gds.GoogleDiskDefault.cfg.FolderID},
	}

	_, err = gds.GoogleDiskDefault.Srv.Files.Create(driveFile).Media(file).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("ошибка загрузки файла: %w", err)
	}

	slog.Info("Файл успешно загружен",
		"filename", filename,
		"fileSize", FormatBytes(fileSize),
		"freeSpaceAfter", FormatBytes(quota.FreeBytes-fileSize),
		slog.String("url", "https://drive.google.com/drive/folders/"+gds.GetIDFolder()),
	)

	return nil
}

// deleteOldCopies удаляет самые старые копии файла, оставляя UploadCopiesCount - 1 копий
func (gds *GoogleDisks) deleteOldCopies(ctx context.Context, filename string) error {
	// Получаем базовое имя файла без пути
	basename := filepath.Base(filename)

	// Получаем список файлов в папке с таким же именем
	query := fmt.Sprintf("'%s' in parents and name = '%s' and trashed = false",
		gds.GoogleDiskDefault.cfg.FolderID, basename)

	files, err := gds.GoogleDiskDefault.Srv.Files.List().Q(query).
		Fields("files(id, name, modifiedTime)").OrderBy("modifiedTime asc").Do()
	if err != nil {
		// Если ошибка "Not found" (404), это нормально - просто нет файлов, выходим без ошибки
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return nil
		}
		return fmt.Errorf("ошибка получения списка файлов: %w", err)
	}

	// Если файлов меньше или равно UploadCopiesCount - 1, ничего не удаляем
	maxCopies := gds.GoogleDiskDefault.cfg.UploadCopiesCount - 1
	if len(files.Files) <= maxCopies {
		return nil
	}

	// Удаляем самые старые файлы, оставляя только maxCopies копий
	filesToDelete := len(files.Files) - maxCopies
	for i := 0; i < filesToDelete; i++ {
		err := gds.GoogleDiskDefault.Srv.Files.Delete(files.Files[i].Id).Context(ctx).Do()
		if err != nil {
			slog.Warn("ошибка удаления файла в google disk", "fileId", files.Files[i].Id, "filename", files.Files[i].Name, "error", err)
		} else {
			slog.Info("удален старый файл в google disk", "filename", files.Files[i].Name, "modifiedTime", files.Files[i].ModifiedTime)
		}
	}

	return nil
}

func (gds *GoogleDisks) GetIDFolder() string {
	return gds.GoogleDiskDefault.cfg.FolderID
}
