package googleupload

import (
	"context"
	"fmt"
)

// StorageQuota содержит информацию о квоте хранилища
type StorageQuota struct {
	TotalBytes  int64 `json:"quotaBytesTotal"`       // Общий размер квоты
	UsedBytes   int64 `json:"quotaBytesUsed"`        // Использованное место
	FreeBytes   int64 `json:"freeBytesRemaining"`    // Свободное место
	UsedInTrash int64 `json:"quotaBytesUsedInTrash"` // Место в корзине
}

// GetStorageQuota получает информацию о квоте хранилища Google Drive
func (gd *GoogleDisk) GetStorageQuota(ctx context.Context) (*StorageQuota, error) {
	about, err := gd.Srv.About.Get().Fields("storageQuota").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о квоте: %w", err)
	}

	quota := &StorageQuota{
		TotalBytes:  about.StorageQuota.Limit,
		UsedBytes:   about.StorageQuota.Usage,
		UsedInTrash: about.StorageQuota.UsageInDriveTrash,
	}

	// Вычисляем свободное место
	quota.FreeBytes = quota.TotalBytes - quota.UsedBytes

	return quota, nil
}

// GetStorageQuotaDetailed получает детальную информацию о квоте с дополнительными полями
func (gd *GoogleDisk) GetStorageQuotaDetailed(ctx context.Context) (*StorageQuota, error) {
	about, err := gd.Srv.About.Get().Fields("storageQuota").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения детальной информации о квоте: %w", err)
	}

	quota := &StorageQuota{
		TotalBytes:  about.StorageQuota.Limit,
		UsedBytes:   about.StorageQuota.Usage,
		UsedInTrash: about.StorageQuota.UsageInDriveTrash,
	}

	// Вычисляем свободное место
	quota.FreeBytes = quota.TotalBytes - quota.UsedBytes

	return quota, nil
}

// HasEnoughSpace проверяет, достаточно ли свободного места для файла указанного размера
func (gd *GoogleDisk) HasEnoughSpace(ctx context.Context, fileSize int64) (bool, *StorageQuota, error) {
	quota, err := gd.GetStorageQuota(ctx)
	if err != nil {
		return false, nil, err
	}

	hasSpace := quota.FreeBytes >= fileSize
	return hasSpace, quota, nil
}

// FormatBytes форматирует размер в байтах в читаемый вид
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f ТБ", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f ГБ", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f МБ", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f КБ", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d байт", bytes)
	}
}
