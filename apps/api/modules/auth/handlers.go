package auth

import (
	"github.com/gofiber/fiber/v3"
)

func GetGoogleAuth(c fiber.Ctx) error {
	google := ConfigGoogle()
	url := google.AuthCodeURL("state")
	return c.Redirect().To(url)
}

func GetGoogleAuthCallback(c fiber.Ctx) error {
	token, error := ConfigGoogle().Exchange(c.Context(), c.FormValue("code"))
	if error != nil {
		panic(error)
	}
	return c.Status(200).JSON(fiber.Map{"token": token.AccessToken, "login": true})
}
