package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Lil-Strudel/glassact-studios/apps/api/database"
	"github.com/Lil-Strudel/glassact-studios/apps/api/router"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()
	database.Migrate("up")

	handler := http.NewServeMux()
	router.SetupRoutes(handler)

	server := http.Server{
		Addr:    ":4100",
		Handler: handler,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Printf("Gracefully shutting down...")
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		log.Printf("Running cleanup tasks...")
		database.Db.Close()

		log.Printf("Successful shutdown!")
		close(idleConnsClosed)
	}()

	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
