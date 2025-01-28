package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/database"
	"github.com/Lil-Strudel/glassact-studios/apps/api/model"
	"github.com/Lil-Strudel/glassact-studios/apps/api/util"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func GetGoogleAuth(c fiber.Ctx) error {
	google := ConfigGoogle()
	url := google.AuthCodeURL("state")
	return c.Redirect().To(url)
}

func GetGoogleAuthCallback(c fiber.Ctx) error {
	token, err := ConfigGoogle().Exchange(c.Context(), c.FormValue("code"))
	if err != nil {
		panic(err)
	}

	userInfo := GetGoogleUserInfo(token.AccessToken)

	var account *model.Account
	var user *model.User

	existingAccount, exists := FindExistingAccount("google", userInfo.ID)

	if exists {
		existingUser, exists := GetUserByID(existingAccount.UserID)

		if !exists {
			panic(fmt.Sprintf("A user with id %v could not be found for account id %v", existingAccount.UserID, existingAccount.ID))
		}

		user = existingUser
		account = existingAccount
	} else {
		existingUser, exists := GetUserByEmail(userInfo.Email)

		if exists {
			newAccount := model.Account{
				UserID:            existingUser.ID,
				Type:              "oidc",
				Provider:          "google",
				ProviderAccountID: userInfo.ID,
				AccessToken:       &token.AccessToken,
				Expires:           &token.Expiry,
			}

			newAcc, err := CreateNewAccount(newAccount)
			if err != nil {
				panic(err)
			}

			user = existingUser
			account = newAcc

		} else {
			newUser := model.User{
				Email:         userInfo.Email,
				EmailVerified: util.Ptr(time.Now()),
				Image:         &userInfo.Picture,
			}
			newAccount := model.Account{
				Type:              "oidc",
				Provider:          "google",
				ProviderAccountID: userInfo.ID,
				AccessToken:       &token.AccessToken,
				Expires:           &token.Expiry,
			}

			newUsr, newAcc, err := CreateNewUser(newUser, newAccount)
			if err != nil {
				panic(err)
			}

			user = newUsr
			account = newAcc
		}

	}

	sess := session.FromContext(c)
	if sess.Fresh() {
		sid := sess.ID()
		uid := user.ID

		sess.Set("uid", uid)
		sess.Set("sid", sid)
		sess.Set("ip", c.IP())
		sess.Set("login", time.Unix(time.Now().Unix(), 0).UTC().String())
		sess.Set("ua", string(c.Request().Header.UserAgent()))

		_, err := database.Db.Exec(context.Background(), `UPDATE sessions SET user_id = $1 WHERE id = $2`, uid, sid)
		if err != nil {
			panic(err)
		}

		fmt.Println("created sess")
	}

	return c.Status(200).JSON(fiber.Map{"user": user, "account": account})
}
