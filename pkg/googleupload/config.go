package googleupload

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"go.yaml.in/yaml/v3"
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
	UploadCopiesCount     int    `yaml:"upload_copies_count" mapstructure:"upload_copies_count" default:"1"`
	FolderID              string `yaml:"folder_id" mapstructure:"folder_id"`
	Enable                bool   `yaml:"enable" mapstructure:"enable" default:"true"`
}

// LoadConfig загружает конфигурацию из YAML файлов
// Если файлы не указаны, использует config.yaml
// Поддерживает значения по умолчанию из тегов default
func LoadConfig(yamlFiles ...string) (*Config, error) {
	// Если файлы не указаны, используем config.yaml по умолчанию
	if len(yamlFiles) == 0 {
		yamlFiles = []string{ConfigFilyDefault}
	}

	cfg := &Config{}

	// Загружаем данные из указанных YAML файлов
	// Последующие файлы перезаписывают значения из предыдущих
	for _, file := range yamlFiles {
		if err := loadYaml(file, cfg); err != nil {
			return nil, fmt.Errorf("ошибка загрузки файла конфигурации %s: %v", file, err)
		}
	}

	// значения по умолчанию из тегов default
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("ошибка установки значений по умолчанию: %v", err)
	}

	// Валидируем конфигурацию Google Drive
	for _, drive := range cfg.ConfigGoogleDrives {
		if drive.Enable {
			if err := drive.Validate(); err != nil {
				return nil, err
			}
		}
	}

	return cfg, nil
}

// loadYaml загружает конфигурацию из YAML файла
func loadYaml(file string, cfg *Config) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}

// Validate проверяет конфигурацию Google Drive
func (c *ConfigGoogleDrive) Validate() error {
	_, err := os.Stat(c.GoogleCredentialsFile)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		url := "https://github.com/san035/google-drive-upload/tree/main/docs/CREDENTIALS_GOOGLE_DRIVE.md"
		_ = openBrowser(url)
		return fmt.Errorf("файл GoogleCredentialsFile не найден: %s. Как его получить: %s", c.GoogleCredentialsFile, url)
	}
	return err
}
