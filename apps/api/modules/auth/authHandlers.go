package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/libs/database"
)

type application struct {
	*app.Application
}

func NewAuthModule(app *app.Application) *application {
	return &application{
		app,
	}
}

func (app *application) HandleGetGoogleAuth(w http.ResponseWriter, r *http.Request) {
	google := app.configGoogle()
	url := google.AuthCodeURL("state")

	http.Redirect(w, r, url, http.StatusFound)
}

func (app *application) HandleGetGoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	token, err := app.configGoogle().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		return
	}

	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		return
	}

	var user *database.User

	existingAccount, found, err := app.Db.Accounts.GetByProvider("google", userInfo.ID)
	if err != nil {
		app.Log.Info("error while finding account")
		return
	}

	if found {
		existingUser, found, err := app.Db.Users.GetByID(existingAccount.UserID)
		if err != nil {
			app.Log.Info("account + error while finding user")
			return
		}

		if !found {
			app.Log.Info("account found but no user")
			return
		}

		user = existingUser
	} else {
		existingUser, found, err := app.Db.Users.GetByEmail(userInfo.Email)
		if err != nil {
			app.Log.Info("no account + error when finding user")
			return
		}

		if !found {
			app.Log.Info("no account + no user")
			return
		}

		newAccount := database.Account{
			UserID:            existingUser.ID,
			Type:              "oidc",
			Provider:          "google",
			ProviderAccountID: userInfo.ID,
		}

		err = app.Db.Accounts.Insert(&newAccount)
		if err != nil {
			app.Log.Info("error creating account")
			return
		}

		user = existingUser
	}

	refreshToken, err := app.Db.Tokens.New(user.ID, 30*24*time.Hour, database.ScopeRefresh)
	if err != nil {
		app.Log.Info("error creating refreshToken")
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
	http.Redirect(w, r, app.Cfg.BaseUrl, http.StatusFound)
}

func (app *application) HandlePostTokenAccess(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			app.WriteJSON(w, r, http.StatusUnauthorized, map[string]any{
				"message": "No refresh token found in cookie",
			})
			return
		default:
			app.Log.Info(err.Error())
			return
		}
	}

	user, found, err := app.Db.Users.GetForToken(database.ScopeRefresh, cookie.Value)
	if err != nil {
		app.Log.Info(err.Error())
		return
	}

	if !found {
		app.Log.Info("not found")
		return
	}

	accessToken, err := app.Db.Tokens.New(user.ID, 2*time.Hour, database.ScopeAccess)
	if err != nil {
		return
	}

	app.WriteJSON(w, r, http.StatusCreated, map[string]any{
		"access_token":     accessToken.Plaintext,
		"access_token_exp": accessToken.Expiry,
	})
}

func (app *application) HandleGetLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			app.WriteJSON(w, r, http.StatusUnauthorized, map[string]any{
				"message": "No refresh token found in cookie",
			})
			return
		default:
			app.Log.Info(err.Error())
			return
		}
	}

	err = app.Db.Tokens.DeleteByPlaintext(database.ScopeRefresh, cookie.Value)
	if err != nil {
		app.Log.Info(err.Error())
		return
	}

	newCookie := http.Cookie{
		Name:   "refresh_token",
		Path:   "/api/auth",
		MaxAge: -1,
	}

	http.SetCookie(w, &newCookie)

	http.Redirect(w, r, app.Cfg.BaseUrl, http.StatusFound)
}
