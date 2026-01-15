package config

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

type Config struct {
	GoogleCredentialsFile string `json:"google_credentials_file"`
	Oauth2Config          *oauth2.Config
	WebCode               string
}

const (
	credentialsFile = "google_credentials.json"
)

func LoadConfig() (*Config, error) {

	data, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}

	oauth2Config, err := google.ConfigFromJSON(data, drive.DriveScope)
	if err != nil {
		return nil, err
	}

	return &Config{Oauth2Config: oauth2Config}, nil
}
