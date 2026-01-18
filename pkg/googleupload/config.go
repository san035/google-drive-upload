package googleupload

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

const ConfigFilyDefault = "config.yaml"

type Config struct {
	OAuthCallbackHostPort string             `yaml:"oauth_callback_host_port" default:"localhost:8080"` // Хост и порт для OAuth callback (по умолчанию "localhost:80909")
	ConfigGoogleDrives    ConfigGoogleDrives `yaml:"config_google_drives"`
}

type ConfigGoogleDrives []*ConfigGoogleDrive

type ConfigGoogleDrive struct {
	Id                    string `yaml:"id" envconfig:"ID"`
	GoogleCredentialsFile string `yaml:"google_credentials_file" envconfig:"GOOGLE_CREDENTIALS_FILE"`
	UploadCopiesCount     int    `yaml:"upload_copies_count" envconfig:"UPLOAD_COPIES_COUNT"`
	FolderID              string `yaml:"folder_id" envconfig:"FOLDER_ID"`
}

// LoadConfig загружает конфигурацию из YAML файлов с помощью cleanenv
// Если файлы не указаны, использует config.yaml
func LoadConfig(yamlFiles ...string) (*Config, error) {
	// Если файлы не указаны, используем config.yaml по умолчанию
	if len(yamlFiles) == 0 {
		yamlFiles = []string{ConfigFilyDefault}
	}

	config := &Config{}

	// Загружаем данные из указанных YAML файлов с помощью cleanenv
	// Последующие файлы перезаписывают значения из предыдущих
	for _, file := range yamlFiles {
		if err := cleanenv.ReadConfig(file, config); err != nil {
			return nil, fmt.Errorf("ошибка чтения файла конфигурации %s: %v", file, err)
		}
	}

	return config, nil
}
