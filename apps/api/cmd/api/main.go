package main

import (
	"log/slog"
	"os"
	"sync"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/config"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules"
	"github.com/Lil-Strudel/glassact-studios/libs/database"
	"github.com/go-playground/validator/v10"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	db, err := database.NewPool(cfg.Db.Dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &app.Application{
		Cfg:      cfg,
		Db:       database.NewModels(db),
		Err:      app.AppError,
		Log:      logger,
		Validate: validator.New(validator.WithRequiredStructEnabled()),
		Wg:       sync.WaitGroup{},
	}

	err = app.Serve(modules.GetRoutes(app))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
