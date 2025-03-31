package main

import (
	"context"
	"encoding/json"
	"errors"
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

func (app *application) handleGetGoogleAuth(w http.ResponseWriter, r *http.Request) {
	google := app.configGoogle()
	url := google.AuthCodeURL("state")

	http.Redirect(w, r, url, http.StatusFound)
}

func (app *application) handleGetGoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	token, err := app.configGoogle().Exchange(context.Background(), r.FormValue("code"))
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
		Path:     "/api/auth",
		Expires:  refreshToken.Expiry,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, app.cfg.baseUrl, http.StatusFound)
}

func (app *application) handlePostTokenAccess(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			app.writeJSON(w, http.StatusUnauthorized, map[string]any{
				"message": "No refresh token found in cookie",
			})
			return
		default:
			app.log.Info(err.Error())
			return
		}
	}

	user, found, err := app.db.Users.GetForToken(database.ScopeRefresh, cookie.Value)
	if err != nil {
		app.log.Info(err.Error())
		return
	}

	if !found {
		app.log.Info("not found")
		return
	}

	accessToken, err := app.db.Tokens.New(user.ID, 2*time.Hour, database.ScopeAccess)
	if err != nil {
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]any{
		"access_token":     accessToken.Plaintext,
		"access_token_exp": accessToken.Expiry,
	})
}

func (app *application) handleGetLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			app.writeJSON(w, http.StatusUnauthorized, map[string]any{
				"message": "No refresh token found in cookie",
			})
			return
		default:
			app.log.Info(err.Error())
			return
		}
	}

	err = app.db.Tokens.DeleteByPlaintext(database.ScopeRefresh, cookie.Value)
	if err != nil {
		app.log.Info(err.Error())
		return
	}

	newCookie := http.Cookie{
		Name:   "refresh_token",
		Path:   "/api/auth",
		MaxAge: -1,
	}

	http.SetCookie(w, &newCookie)

	http.Redirect(w, r, app.cfg.baseUrl, http.StatusFound)
}

func (app *application) handleGetUserSelf(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	app.writeJSON(w, http.StatusOK, user)
}
