package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/model"
	"github.com/Lil-Strudel/glassact-studios/apps/api/util"
)

func GetGoogleAuth(w http.ResponseWriter, req *http.Request) {
	google := ConfigGoogle()
	url := google.AuthCodeURL("state")

	http.Redirect(w, req, url, http.StatusFound)
}

func GetGoogleAuthCallback(w http.ResponseWriter, req *http.Request) {
	token, err := ConfigGoogle().Exchange(context.Background(), req.FormValue("code"))
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	type Response struct {
		User    *model.User    `json:"user"`
		Account *model.Account `json:"account"`
	}
	json.NewEncoder(w).Encode(Response{
		User:    user,
		Account: account,
	})
}
