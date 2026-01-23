package googleupload

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// GetToken возвращает токен доступа (из кэша, обновляет или запрашивает новый)
func (gd *GoogleDisk) GetToken(config *oauth2.Config) (*oauth2.Token, error) {
	l := slog.With("idDisk", gd.cfg.Id)
	tokenFile := strings.TrimSuffix(gd.cfg.GoogleCredentialsFile, filepath.Ext(gd.cfg.GoogleCredentialsFile)) + "_token.json"
	token, err := LoadToken(tokenFile)
	if err == nil {
		// Если токен валиден - возвращаем сразу
		if token.Valid() {
			return token, nil
		}

		// Если токен истёк, но есть refresh token - пробуем обновить
		if token.RefreshToken != "" {
			l.Info("Токен истёк, пробуем обновить через refresh token")
			newToken, refreshErr := config.TokenSource(context.Background(), token).Token()
			if refreshErr == nil {
				// Сохраняем обновлённый токен
				saveErr := SaveToken(tokenFile, newToken)
				if saveErr == nil {
					l.Info("Token update and save")
					return newToken, nil
				}
				l.Warn("Не удалось сохранить обновлённый токен", "error", saveErr)
				return newToken, nil
			}
			l.Warn("Не удалось обновить токен", "error", refreshErr)
		}
	}

	// Если токена нет, он истёк без refresh token или ошибка загрузки - запрашиваем новый
	code, err := gd.getCodeAuth(config)
	if err != nil {
		return nil, err
	}

	// Обмен кода на токен
	newToken, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("ошибка обмена кода на токен: %w", err)
	}

	// Сохраняем новый токен
	if err := SaveToken(tokenFile, newToken); err != nil {
		slog.Warn("Не удалось сохранить токен", "error", err)
	}

	return newToken, nil
}

func (gd *GoogleDisk) getCodeAuth(config *oauth2.Config) (string, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// HTTP сервер для приёма редиректа
	// Извлекаем хост и порт из config.RedirectURL
	redirectURL, err := url.Parse(config.RedirectURL)
	if err != nil {
		return "", err
	}

	// Канал для получения кода авторизации
	codeChan := make(chan string, 1)

	server := &http.Server{Addr: redirectURL.Host}
	http.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if state != "state-token" {
			http.Error(w, "Неверный state параметр", http.StatusBadRequest)
			return
		}

		if code == "" {
			http.Error(w, "Код не получен", http.StatusBadRequest)
			return
		}

		// Отправляем код в канал
		codeChan <- code

		// Возвращаем страницу успеха
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<body>
			Authorization successful
			</body>
			</html>
		`)
	})

	// Запускаем сервер
	go func() {
		slog.Info("Ожидание авторизации", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ошибка HTTP сервера", "error", err)
		}
	}()

	// Открываем браузер с ссылкой авторизации
	slog.Info("Открываю браузер для авторизации", "url", authURL)
	if err := openBrowser(authURL); err != nil {
		slog.Warn("Не удалось открыть браузер, скопируйте ссылку вручную", "url", authURL)
	}

	// Ждём получение кода с таймаутом 5 минут
	var code string
	select {
	case code = <-codeChan:
		slog.Info("Получен код авторизации")
		_ = server.Shutdown(context.Background())
	case <-time.After(5 * time.Minute):
		_ = server.Shutdown(context.Background())
		return "", fmt.Errorf("время ожидания авторизации истекло (5 минут)")
	}
	return code, nil
}

// LoadToken загружает токен из JSON файла
func LoadToken(fileToken string) (*oauth2.Token, error) {
	data, err := DecryptFile(fileToken)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{}
	if err := json.Unmarshal(data, token); err != nil {
		return nil, err
	}

	slog.Debug("Успешно загружен токен из файла", "expiry", token.Expiry, "file", fileToken)
	return token, nil
}

// SaveToken сохраняет токен в JSON файл
func SaveToken(fileToken string, token *oauth2.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	err = EncryptContentAndSaveToFile(fileToken, data)
	return err
}

// openBrowser открывает URL в браузере по умолчанию
func openBrowser(urlStr string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Используем rundll32 для надёжного открытия URL на Windows
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", urlStr)
	case "darwin":
		cmd = exec.Command("open", urlStr)
	case "linux":
		// Пробуем различные команды для Linux
		for _, browser := range []string{"xdg-open", "firefox", "google-chrome", "chromium"} {
			if _, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(browser, urlStr)
				break
			}
		}
		if cmd == nil {
			return fmt.Errorf("не найден браузер для открытия URL")
		}
	default:
		return fmt.Errorf("неподдерживаемая ОС: %s", runtime.GOOS)
	}

	return cmd.Start()
}
