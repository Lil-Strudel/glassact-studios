package main

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/config"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, stdb, err := data.NewPool(cfg.Db.Dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.S3.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.S3.AccessKeyID,
			cfg.S3.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		logger.Error("failed to load AWS config", "error", err.Error())
		os.Exit(1)
	}
	s3Client := s3.NewFromConfig(awsCfg)

	app := &app.Application{
		Cfg:      cfg,
		Db:       data.NewModels(db, stdb),
		Err:      app.AppError,
		Log:      logger,
		Validate: validator.New(validator.WithRequiredStructEnabled()),
		Wg:       sync.WaitGroup{},
		S3:       s3Client,
	}

	err = app.Serve(modules.GetRoutes(app))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
