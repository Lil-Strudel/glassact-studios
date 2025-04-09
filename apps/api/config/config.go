package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env     string
	Port    int
	BaseUrl string
	Db      struct {
		Dsn string
	}
	Google struct {
		ClientId     string
		ClientSecret string
		RedirectUrl  string
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
	cfg.BaseUrl = os.Getenv("BASE_URL")

	cfg.Db.Dsn = os.Getenv("DATABASE_DSN")

	cfg.Google.ClientId = os.Getenv("GOOGLE_CLIENT_ID")
	cfg.Google.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	cfg.Google.RedirectUrl = os.Getenv("GOOGLE_REDIRECT_URL")

	return &cfg, nil
}
