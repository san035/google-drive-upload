package configgoogledrive

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ConfigGoogleDrives ConfigGoogleDrives `yaml:"config_google_drives"`
}

type ConfigGoogleDrives []*ConfigGoogleDrive

type ConfigGoogleDrive struct {
	Id                    string `yaml:"id" envconfig:"ID"`
	GoogleCredentialsFile string `yaml:"google_credentials_file" envconfig:"GOOGLE_CREDENTIALS_FILE"`
	UploadCopiesCount     int    `yaml:"upload_copies_count" envconfig:"UPLOAD_COPIES_COUNT"`
	FolderID              string `yaml:"folder_id" envconfig:"FOLDER_ID"`
}

// LoadConfig загружает конфигурацию из YAML файлов
// Если файлы не указаны, использует config.yaml
func LoadConfig(yamlFiles ...string) (*Config, error) {
	// Если файлы не указаны, используем config.yaml по умолчанию
	if len(yamlFiles) == 0 {
		yamlFiles = []string{"config.yaml"}
	}

	config := &Config{}
	var configData []byte

	// Загружаем данные из указанных YAML файлов
	for _, file := range yamlFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("ошибка чтения файла конфигурации %s: %v", file, err)
		}
		configData = append(configData, data...)
	}

	// Парсим YAML данные
	if err := yaml.Unmarshal(configData, config); err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML конфигурации: %v", err)
	}

	return config, nil
}
