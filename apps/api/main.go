package main

import (
	"log"

	"github.com/Lil-Strudel/glassact-studios/apps/api/database"
	"github.com/Lil-Strudel/glassact-studios/apps/api/router"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recoverer.New())

	router.SetupRoutes(app)

	log.Fatal(app.Listen(":4100"))
}
