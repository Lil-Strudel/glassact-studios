package auth

import "github.com/gofiber/fiber/v3"

func SetupRoutes(api fiber.Router) {
	auth := api.Group("/auth")

	auth.Get("/google", GetGoogleAuth)
	auth.Get("/google/callback", GetGoogleAuthCallback)
}
