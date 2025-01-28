package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Lil-Strudel/glassact-studios/apps/api/database"
	"github.com/Lil-Strudel/glassact-studios/apps/api/router"
	"github.com/Lil-Strudel/glassact-studios/apps/api/sessionStore"
	"github.com/gofiber/fiber/v3"
	// "github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/session"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()
	database.Migrate("up")

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recoverer.New())

	storage := sessionStore.New(sessionStore.Config{
		DB:    database.Db,
		Table: "sessions",
	})
	sessionMiddleware, _ := session.NewWithStore(session.Config{
		Storage: storage,
	})

	app.Use(sessionMiddleware)
	// app.Use(csrf.New(csrf.Config{}))

	router.SetupRoutes(app)

	go func() {
		err := app.Listen(":4100", fiber.ListenConfig{
			DisableStartupMessage: true,
		})
		if err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	_ = <-c

	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	fmt.Println("Running cleanup tasks...")
	database.Db.Close()

	fmt.Println("Fiber was successful shutdown.")
}
