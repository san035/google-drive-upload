package googleupload

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// UploadFile upload file to Google Drive
// example googleupload.UploadFile(ctx, "test.zip, UseIDDisk("1"))
func (gds *GoogleDisks) UploadFile(ctx context.Context, filename string, idDisk string) error {
	gd, err := gds.findGDById(idDisk)
	if err != nil {
		return err
	}

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

	_, err = gd.Srv.Files.Create(driveFile).Media(file).Context(ctx).Do()
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

func (gds *GoogleDisks) findGDById(idDisk string) (*GoogleDisk, error) {
	var gd *GoogleDisk
	if len(idDisk) == 0 {
		gd = gds.GoogleDiskDefault
	} else {
		for _, v := range gds.ListGoogleDisk {
			if v.cfg.Id == idDisk {
				gd = v
				break
			}
		}
		if gd == nil {
			return nil, errors.New("unknow ID disk")
		}
	}
	return gd, nil
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
