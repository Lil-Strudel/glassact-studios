package cat

import "github.com/gofiber/fiber/v3"

func SetupRoutes(api fiber.Router) {
	cat := api.Group("/cat")

	cat.Get("/", GetCats)
	cat.Post("/", PostCat)
}
