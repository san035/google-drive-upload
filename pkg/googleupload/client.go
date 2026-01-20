package googleupload

import (
	"context"
	"errors"
	"log/slog"
	"os"

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
	cfg *ConfigGoogleDrive
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
		if !cfg.Enable {
			continue
		}

		data, err := os.ReadFile(cfg.GoogleCredentialsFile)
		if err != nil {
			// Проверяем, если файл не найден (windows ERROR_FILE_NOT_FOUND = 2, syscall ENOENT)
			if os.IsNotExist(err) {
				slog.Error("Файл не найден. Как его получить: https://github.com/san035/google-drive-upload/tree/main/docs/CREDENTIALS_GOOGLE_DRIVE.md", "filename", cfg.GoogleCredentialsFile, "error", err)
				panic(err)
			}
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

	if len(listGoogleDisk) == 0 {
		return nil, errors.New("no set config_google_drives")
	}

	gds := GoogleDisks{
		GoogleDiskDefault: gdDefault,
		ListGoogleDisk:    listGoogleDisk,
	}
	return &gds, nil
}

func (gds *GoogleDisks) GetIDFolder() string {
	return gds.GoogleDiskDefault.cfg.FolderID
}
