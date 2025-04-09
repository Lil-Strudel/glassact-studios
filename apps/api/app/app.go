package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/config"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/go-playground/validator/v10"
)

type Application struct {
	Cfg      *config.Config
	Db       data.Models
	Err      appError
	Log      *slog.Logger
	Validate *validator.Validate
	Wg       sync.WaitGroup
}

func (app *Application) Serve(routes http.Handler) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Cfg.Port),
		Handler:      routes,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.Log.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.Log.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.Log.Info("closing db pool")
		app.Db.Pool.Close()

		app.Log.Info("completing background tasks", "addr", srv.Addr)
		app.Wg.Wait()

		shutdownError <- nil
	}()

	app.Log.Info("starting server", "addr", srv.Addr, "env", app.Cfg.Env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Log.Info("stopped server", "addr", srv.Addr)

	return nil
}
