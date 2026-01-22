package googleupload

import (
	"context"
	"errors"

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

		data, err := DecryptFile(cfg.GoogleCredentialsFile)
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

	if len(listGoogleDisk) == 0 {
		return nil, errors.New("no set config_google_drives")
	}

	gds := GoogleDisks{
		GoogleDiskDefault: gdDefault,
		ListGoogleDisk:    listGoogleDisk,
	}
	return &gds, nil
}

func (gd *GoogleDisk) GetUrlFile() string {
	if gd.cfg.FolderID == "" {
		return `https://drive.google.com/drive/my-drive`
	}

	return "https://drive.google.com/drive/folders/" + gd.cfg.FolderID
}
