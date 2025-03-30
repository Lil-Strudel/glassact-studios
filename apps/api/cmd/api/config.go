package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type config struct {
	env     string
	port    int
	baseUrl string
	db      struct {
		dsn string
	}
	google struct {
		clientId     string
		clientSecret string
		redirectUrl  string
	}
}

func getConfig() (*config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	var cfg config

	cfg.env = os.Getenv("ENV")

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return nil, err
	}

	cfg.port = port
	cfg.baseUrl = os.Getenv("BASE_URL")

	cfg.db.dsn = os.Getenv("DATABASE_DSN")

	cfg.google.clientId = os.Getenv("GOOGLE_CLIENT_ID")
	cfg.google.clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	cfg.google.redirectUrl = os.Getenv("GOOGLE_REDIRECT_URLy")

	return &cfg, nil
}
