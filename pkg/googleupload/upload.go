package googleupload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// progressReader обёртка для Reader с отслеживанием прогресса загрузки
type progressReader struct {
	reader        io.Reader
	totalSize     int64
	uploadedBytes int64
}

// Read реализует io.Reader с подсчётом прочитанных байт
func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.uploadedBytes += int64(n)
	return n, err
}

// Progress возвращает текущий прогресс в байтах
func (pr *progressReader) Progress() int64 {
	return pr.uploadedBytes
}

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

	// Умная очистка корзины: очищаем только если не хватает места
	if err := gds.smartClearTrash(ctx, fileSize); err != nil {
		return err
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
		Name: filepath.Base(filename),
	}
	// Если FolderID указан, загружаем в папку, иначе - в корень диска
	if gd.cfg.FolderID != "" {
		driveFile.Parents = []string{gd.cfg.FolderID}
	}

	// Создаём progressReader для отслеживания прогресса загрузки
	pr := &progressReader{
		reader:        file,
		totalSize:     fileSize,
		uploadedBytes: 0,
	}

	// Запускаем горутину для логирования прогресса раз в минуту
	progressChan := make(chan int64, 1)
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				uploaded := pr.Progress()
				percent := float64(uploaded) / float64(fileSize) * 100
				slog.Info("progress upload in Google Drive",
					"filename", filename,
					"uploaded", FormatBytes(uploaded),
					"total", FormatBytes(fileSize),
					"progress", fmt.Sprintf("%.2f%%", percent),
				)
				progressChan <- uploaded
			}
		}
	}()

	_, err = gd.Srv.Files.Create(driveFile).Media(pr).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("error upload file: %w", err)
	}

	slog.Info("Success upload file",
		"filename", filename,
		"idDisk", idDisk,
		"fileSize", FormatBytes(fileSize),
		slog.String("url", gd.GetUrlFolderFile()),
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
			return nil, errors.New("unknow ID disk: " + idDisk)
		}
	}
	return gd, nil
}

// emptyTrash очищает корзину Google Drive (безвозвратно удаляет файлы из корзины)
// Удаляет файлы начиная со старых, пока не освободит至少 clearSize байт
func (gds *GoogleDisks) emptyTrash(ctx context.Context, clearSize int64) error {
	query := "'me' in owners and trashed = true"

	files, err := gds.GoogleDiskDefault.Srv.Files.List().Q(query).
		Fields("files(id, name, size, trashedTime, createdTime)").
		OrderBy("trashedTime asc").
		Do()
	if err != nil {
		return fmt.Errorf("ошибка получения списка файлов в корзине: %w", err)
	}

	if len(files.Files) == 0 {
		return nil
	}

	var clearedSize int64
	for _, file := range files.Files {
		// Проверяем, достаточно ли уже освобождено места
		if clearedSize >= clearSize && clearSize > 0 {
			slog.Info("достигнут лимит освобождаемого места", "clearedSize", clearedSize, "clearSize", clearSize)
			break
		}

		err := gds.GoogleDiskDefault.Srv.Files.Delete(file.Id).Context(ctx).Do()
		if err != nil {
			slog.Warn("ошибка удаления файла из корзины", "filename", file.Name, "createdTime", file.CreatedTime, "error", err)
			continue
		}

		slog.Info("файл безвозвратно удалён из корзины", "filename", file.Name, "createdTime", file.CreatedTime, "size", file.Size)
		clearedSize += file.Size
	}

	slog.Info("очистка корзины завершена", "clearedSize", clearedSize, "clearSize", clearSize)
	return nil
}

// smartClearTrash очищает корзину Google Drive только когда не хватает места для загрузки файла
func (gds *GoogleDisks) smartClearTrash(ctx context.Context, fileSize int64) error {
	// Проверяем наличие свободного места
	hasSpace, quota, err := gds.GoogleDiskDefault.HasEnoughSpace(ctx, fileSize)
	if err != nil {
		return fmt.Errorf("ошибка проверки свободного места: %w", err)
	}

	// Если места достаточно, не очищаем корзину
	if hasSpace {
		return nil
	}

	slog.Warn("недостаточно места на Google Drive, пробуем очистить корзину",
		"required", FormatBytes(fileSize),
		"free", FormatBytes(quota.FreeBytes),
		"total", FormatBytes(quota.TotalBytes),
	)

	// Очищаем корзину, освобождая至少 fileSize места
	if err := gds.emptyTrash(ctx, fileSize); err != nil {
		slog.Warn("ошибка очистки корзины Google Disk", "error", err)
		// Не прерываем процесс, пробуем проверить место снова
	}

	// Проверяем наличие свободного места после очистки корзины
	hasSpace, quota, err = gds.GoogleDiskDefault.HasEnoughSpace(ctx, fileSize)
	if err != nil {
		return fmt.Errorf("ошибка проверки свободного места после очистки корзины: %w", err)
	}

	if !hasSpace {
		return fmt.Errorf("недостаточно свободного места на Google Drive даже после очистки корзины. Требуется: %s, свободно: %s (всего: %s, используется: %s)",
			FormatBytes(fileSize), FormatBytes(quota.FreeBytes), FormatBytes(quota.TotalBytes), FormatBytes(quota.UsedBytes))
	}

	slog.Info("корзина очищена, места достаточно для загрузки",
		"required", FormatBytes(fileSize),
		"free", FormatBytes(quota.FreeBytes),
	)

	return nil
}

// deleteOldCopies удаляет самые старые копии файла, оставляя UploadCopiesCount - 1 копий
func (gds *GoogleDisks) deleteOldCopies(ctx context.Context, filename string) error {
	// Получаем базовое имя файла без пути
	basename := filepath.Base(filename)

	// Получаем список файлов в папке с таким же именем
	var query string
	if gds.GoogleDiskDefault.cfg.FolderID != "" {
		query = fmt.Sprintf("'%s' in parents and name = '%s' and trashed = false",
			gds.GoogleDiskDefault.cfg.FolderID, basename)
	} else {
		// Если FolderID пустой, ищем файлы в корне диска (без родителя)
		query = fmt.Sprintf("name = '%s' and trashed = false and 'root' in parents",
			basename)
	}

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
