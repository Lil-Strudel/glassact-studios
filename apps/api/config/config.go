package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env     string
	Port    int
	BaseURL string
	Db      struct {
		Dsn string
	}
	AuthSecret string
	Google     struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	Microsoft struct {
		ClientID     string
		ClientSecret string
		RedirectURL  string
	}
	Smtp struct {
		Host     string
		Port     int
		Username string
		Password string
	}
}

func GetConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	var cfg Config

	cfg.Env = os.Getenv("ENV")

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return nil, err
	}

	cfg.Port = port
	cfg.BaseURL = os.Getenv("BASE_URL")

	cfg.Db.Dsn = os.Getenv("DATABASE_DSN")

	cfg.AuthSecret = os.Getenv("AUTH_SECRET")

	cfg.Google.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
	cfg.Google.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	cfg.Google.RedirectURL = os.Getenv("GOOGLE_REDIRECT_URL")

	cfg.Microsoft.ClientID = os.Getenv("MICROSOFT_CLIENT_ID")
	cfg.Microsoft.ClientSecret = os.Getenv("MICROSOFT_CLIENT_SECRET")
	cfg.Microsoft.RedirectURL = os.Getenv("MICROSOFT_REDIRECT_URL")

	stmpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, err
	}

	cfg.Smtp.Host = os.Getenv("SMTP_HOST")
	cfg.Smtp.Port = stmpPort
	cfg.Smtp.Username = os.Getenv("SMTP_USERNAME")
	cfg.Smtp.Password = os.Getenv("SMTP_PASSWORD")

	return &cfg, nil
}
