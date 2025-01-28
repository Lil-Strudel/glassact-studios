package router

import (
	"github.com/Lil-Strudel/glassact-studios/apps/api/module/auth"
	"github.com/Lil-Strudel/glassact-studios/apps/api/module/cat"
	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello World!",
		})
	})

	auth.SetupRoutes(api)
	cat.SetupRoutes(api)
}
