package main

import (
	"log/slog"
	"os"

	"github.com/Lil-Strudel/glassact-studios/libs/database"
	"github.com/go-playground/validator/v10"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := getConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := database.NewPool(cfg.db.dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	app := &application{
		cfg:      cfg,
		db:       database.NewModels(db),
		log:      logger,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
