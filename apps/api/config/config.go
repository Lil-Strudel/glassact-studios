package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Env     string `validate:"required"`
	Port    int    `validate:"required,min=1,max=65535"`
	BaseURL string `validate:"required,url"`
	Db      struct {
		Dsn string `validate:"required"`
	}
	AuthSecret string `validate:"required,min=32"`
	Google     struct {
		ClientID     string `validate:"required"`
		ClientSecret string `validate:"required"`
		RedirectURL  string `validate:"required,url"`
	}
	Microsoft struct {
		ClientID     string `validate:"required"`
		ClientSecret string `validate:"required"`
		RedirectURL  string `validate:"required,url"`
	}
	Smtp struct {
		Host     string `validate:"required"`
		Port     int    `validate:"required,min=1,max=65535"`
		Username string `validate:"required"`
		Password string `validate:"required"`
	}
	S3 struct {
		Bucket          string `validate:"required"`
		Region          string `validate:"required"`
		AccessKeyID     string `validate:"required"`
		SecretAccessKey string `validate:"required"`
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

	smtpPortStr := os.Getenv("SMTP_PORT")
	if smtpPortStr != "" {
		smtpPort, err := strconv.Atoi(smtpPortStr)
		if err != nil {
			return nil, err
		}
		cfg.Smtp.Port = smtpPort
	}

	cfg.Smtp.Host = os.Getenv("SMTP_HOST")
	cfg.Smtp.Username = os.Getenv("SMTP_USERNAME")
	cfg.Smtp.Password = os.Getenv("SMTP_PASSWORD")

	cfg.S3.Bucket = os.Getenv("S3_BUCKET_NAME")
	cfg.S3.Region = os.Getenv("AWS_REGION")
	cfg.S3.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	cfg.S3.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
