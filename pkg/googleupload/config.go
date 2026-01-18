package googleupload

import (
	"fmt"

	"github.com/chinayin/gox/config"
)

const ConfigFilyDefault = "config.yaml"

type Config struct {
	OAuthCallbackHostPort string             `yaml:"oauth_callback_host_port" mapstructure:"oauth_callback_host_port" default:"localhost:8080"` // Хост и порт для OAuth callback (по умолчанию "localhost:8080")
	ConfigGoogleDrives    ConfigGoogleDrives `yaml:"config_google_drives" mapstructure:"config_google_drives"`
}

type ConfigGoogleDrives []*ConfigGoogleDrive

type ConfigGoogleDrive struct {
	Id                    string `yaml:"id" mapstructure:"id" default:"0"`
	GoogleCredentialsFile string `yaml:"google_credentials_file" mapstructure:"google_credentials_file" default:"google_credentials.json"`
	UploadCopiesCount     int    `yaml:"upload_copies_count" mapstructure:"upload_copies_count" default:"3"`
	FolderID              string `yaml:"folder_id" mapstructure:"folder_id"`
}

// LoadConfig загружает конфигурацию из YAML файлов с помощью gox/config
// Если файлы не указаны, использует config.yaml
// Поддерживает значения по умолчанию из тегов default и переменные окружения
func LoadConfig(yamlFiles ...string) (*Config, error) {
	// Если файлы не указаны, используем config.yaml по умолчанию
	if len(yamlFiles) == 0 {
		yamlFiles = []string{ConfigFilyDefault}
	}

	// Создаём загрузчик конфигурации
	loader := config.NewLoader()

	cfg := &Config{}

	// Загружаем данные из указанных YAML файлов
	// gox/config автоматически применяет значения по умолчанию и переменные окружения
	// Последующие файлы перезаписывают значения из предыдущих
	for _, file := range yamlFiles {
		if err := loader.Load(file, cfg); err != nil {
			return nil, fmt.Errorf("ошибка загрузки файла конфигурации %s: %v", file, err)
		}
	}

	return cfg, nil
}
