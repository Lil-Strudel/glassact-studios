package auth

import (
	"fmt"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/model"
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/fiber/v3"
)

func GetGoogleAuth(c fiber.Ctx) error {
	google := ConfigGoogle()
	url := google.AuthCodeURL("state")
	return c.Redirect().To(url)
}

func newUser(name string, email string, emailVerified time.Time, image string) model.User {
	return model.User{
		Name:          &name,
		Email:         email,
		EmailVerified: &emailVerified,
		Image:         &image,
	}
}

func GetGoogleAuthCallback(c fiber.Ctx) error {
	token, error := ConfigGoogle().Exchange(c.Context(), c.FormValue("code"))
	if error != nil {
		panic(error)
	}

	userInfo := GetGoogleUserInfo(token.AccessToken)

	account, exists := FindExistingAccount("google", userInfo.ID)
	if !exists {
		user, exists := GetUserByEmail(userInfo.Email)

		if exists {
			fmt.Println(user)
			// add new account to user and link
		} else {
			// create new account
			user := newUser("akjsdfkjsd", userInfo.Email, time.Now(), userInfo.Picture)
			account := model.Account{}
			err := CreateNewUser(user, account)
			fmt.Println(err)

			return c.Status(404).SendString("Doesnt Exist")
		}

	}

	spew.Print(account)

	return c.Status(200).JSON(fiber.Map{"message": "Congrats you signed in"})
}
