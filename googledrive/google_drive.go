package googledrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

type GoogleDisk struct {
	Srv *drive.Service
}

// NewDriveService создаёт новый сервис Drive API
func NewDriveService(ctx context.Context, config *oauth2.Config, token *oauth2.Token) (*GoogleDisk, error) {
	client := config.Client(ctx, token)
	srv, err := drive.New(client)
	if err != nil {
		return nil, err
	}
	gd := GoogleDisk{
		Srv: srv,
	}
	return &gd, nil
}

// GetToken возвращает токен доступа (из кэша, обновляет или запрашивает новый)
func GetToken(config *oauth2.Config, tokenFile string) (*oauth2.Token, error) {
	token, err := LoadToken(tokenFile)
	if err == nil {
		// Если токен валиден - возвращаем сразу
		if token.Valid() {
			return token, nil
		}

		// Если токен истёк, но есть refresh token - пробуем обновить
		if token.RefreshToken != "" {
			slog.Info("Токен истёк, пробуем обновить через refresh token")
			newToken, refreshErr := config.TokenSource(context.Background(), token).Token()
			if refreshErr == nil {
				// Сохраняем обновлённый токен
				saveErr := SaveToken(tokenFile, newToken)
				if saveErr == nil {
					slog.Info("Токен успешно обновлён и сохранён")
					return newToken, nil
				}
				slog.Warn("Не удалось сохранить обновлённый токен", "error", saveErr)
				return newToken, nil
			}
			slog.Warn("Не удалось обновить токен", "error", refreshErr)
		}
	}

	// Если токена нет, он истёк без refresh token или ошибка загрузки - запрашиваем новый
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	slog.Info("Перейдите по ссылке для авторизации", "url", authURL)

	fmt.Print("Вставьте полученный код: ")
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("ошибка чтения кода: %w", err)
	}

	newToken, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	// Сохраняем новый токен
	if err := SaveToken(tokenFile, newToken); err != nil {
		slog.Warn("Не удалось сохранить токен", "error", err)
	}

	return newToken, nil
}

// LoadToken загружает токен из JSON файла
func LoadToken(path string) (*oauth2.Token, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{}
	if err := json.Unmarshal(data, token); err != nil {
		return nil, err
	}

	slog.Info("Access токен истекает", "expiry", token.Expiry)
	return token, nil
}

// SaveToken сохраняет токен в JSON файл
func SaveToken(path string, token *oauth2.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// UploadFile загружает файл на Google Drive
func (gd *GoogleDisk) UploadFile(filename, folderID string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	driveFile := &drive.File{
		Name:    filename,
		Parents: []string{folderID},
	}

	_, err = gd.Srv.Files.Create(driveFile).Media(file).Context(context.Background()).Do()
	return err
}
