package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/database"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (app *application) handleNotFound(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"message": "route not found",
	}

	app.writeJSON(w, http.StatusNotFound, data)
}

func (app *application) configGoogle() *oauth2.Config {
	oauth := &oauth2.Config{
		ClientID:     app.cfg.google.clientId,
		ClientSecret: app.cfg.google.clientSecret,
		RedirectURL:  app.cfg.google.redirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
	return oauth
}

type GoogleInfoResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Verified bool   `json:"verified_email"`
	Picture  string `json:"picture"`
}

func getGoogleUserInfo(token string) (*GoogleInfoResponse, error) {
	reqURL, err := url.Parse("https://www.googleapis.com/oauth2/v1/userinfo")
	if err != nil {
		return nil, err
	}

	ptoken := fmt.Sprintf("Bearer %s", token)
	res := &http.Request{
		Method: "GET",
		URL:    reqURL,
		Header: map[string][]string{
			"Authorization": {ptoken},
		},
	}
	req, err := http.DefaultClient.Do(res)
	if err != nil {
		return nil, err
	}

	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var data GoogleInfoResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (app *application) handleGetGoogleAuth(w http.ResponseWriter, req *http.Request) {
	google := app.configGoogle()
	url := google.AuthCodeURL("state")

	http.Redirect(w, req, url, http.StatusFound)
}

func (app *application) handleGetGoogleAuthCallback(w http.ResponseWriter, req *http.Request) {
	token, err := app.configGoogle().Exchange(context.Background(), req.FormValue("code"))
	if err != nil {
		return
	}

	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		return
	}

	var user *database.User

	existingAccount, found, err := app.db.Accounts.GetByProvider("google", userInfo.ID)
	if err != nil {
		app.log.Info("error while finding account")
		return
	}

	if found {
		existingUser, found, err := app.db.Users.GetByID(existingAccount.UserID)
		if err != nil {
			app.log.Info("account + error while finding user")
			return
		}

		if !found {
			app.log.Info("account found but no user")
			return
		}

		user = existingUser
	} else {
		existingUser, found, err := app.db.Users.GetByEmail(userInfo.Email)
		if err != nil {
			app.log.Info("no account + error when finding user")
			return
		}

		if !found {
			app.log.Info("no account + no user")
			return
		}

		newAccount := database.Account{
			UserID:            existingUser.ID,
			Type:              "oidc",
			Provider:          "google",
			ProviderAccountID: userInfo.ID,
		}

		err = app.db.Accounts.Insert(&newAccount)
		if err != nil {
			app.log.Info("error creating account")
			return
		}

		user = existingUser
	}

	refreshToken, err := app.db.Tokens.New(user.ID, 30*24*time.Hour, database.ScopeRefresh)
	if err != nil {
		app.log.Info("error creating refreshToken")
		return
	}

	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken.Plaintext,
		Path:     "/",
		Expires:  refreshToken.Expiry,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, req, app.cfg.baseUrl, http.StatusFound)
}
