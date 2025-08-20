package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type authModule struct {
	*app.Application
}

func NewAuthModule(app *app.Application) *authModule {
	return &authModule{
		app,
	}
}

func (authModule *authModule) HandleGetGoogleAuth(w http.ResponseWriter, r *http.Request) {
	google := authModule.configGoogle()
	url := google.AuthCodeURL("state")

	http.Redirect(w, r, url, http.StatusFound)
}

func (authModule *authModule) HandleGetGoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	token, err := authModule.configGoogle().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	user, found, err := authModule.getUserFromProvider(userInfo.Email, "google", userInfo.ID)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	if !found {
		authModule.WriteError(w, r, authModule.Err.AccountNotFound, nil)
		return
	}

	authModule.login(user.ID, w)
	http.Redirect(w, r, authModule.Cfg.BaseUrl, http.StatusFound)
}

func (authModule *authModule) HandleGetMicrosoftAuth(w http.ResponseWriter, r *http.Request) {
	microsoft := authModule.configMicrosoft()
	url := microsoft.AuthCodeURL("state")

	http.Redirect(w, r, url, http.StatusFound)
}

func (authModule *authModule) HandleGetMicrosoftAuthCallback(w http.ResponseWriter, r *http.Request) {
	token, err := authModule.configMicrosoft().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	userInfo, err := getMicrosoftUserInfo(token.AccessToken)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	user, found, err := authModule.getUserFromProvider(userInfo.Email, "microsoft", userInfo.Sub)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	if !found {
		authModule.WriteError(w, r, authModule.Err.AccountNotFound, nil)
		return
	}

	authModule.login(user.ID, w)
	http.Redirect(w, r, authModule.Cfg.BaseUrl, http.StatusFound)
}

func (authModule *authModule) HandlePostMagicLinkAuth(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email" validate:"required,email"`
	}

	err := authModule.ReadJSONBody(w, r, &body)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.BadRequest, err)
		return
	}

	user, found, err := authModule.Db.Users.GetByEmail(body.Email)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}
	if !found {
		authModule.WriteError(w, r, authModule.Err.AccountNotFound, nil)
		return
	}

	loginToken, err := authModule.Db.Tokens.New(user.ID, 2*time.Hour, data.ScopeLogin)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	authModule.emailMagicLink(body.Email, loginToken.Plaintext)
	authModule.WriteJSON(w, r, http.StatusNoContent, nil)
}

func (authModule *authModule) HandleGetMagicLinkCallback(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	token := qs.Get("token")

	user, found, err := authModule.Db.Users.GetForToken(data.ScopeLogin, token)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	if !found {
		authModule.WriteError(w, r, authModule.Err.ServerError, nil)
		return
	}

	authModule.login(user.ID, w)
	http.Redirect(w, r, authModule.Cfg.BaseUrl, http.StatusFound)
}

func (authModule *authModule) HandlePostTokenAccess(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			authModule.WriteError(w, r, authModule.Err.MissingRefreshToken, err)
			return
		default:
			authModule.WriteError(w, r, authModule.Err.ServerError, err)
			return
		}
	}

	user, found, err := authModule.Db.Users.GetForToken(data.ScopeRefresh, cookie.Value)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	if !found {
		authModule.WriteError(w, r, authModule.Err.AccountNotFound, nil)
		return
	}

	accessToken, err := authModule.Db.Tokens.New(user.ID, 2*time.Hour, data.ScopeAccess)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	authModule.WriteJSON(w, r, http.StatusCreated, map[string]any{
		"access_token":     accessToken.Plaintext,
		"access_token_exp": accessToken.Expiry,
	})
}

func (authModule *authModule) HandleGetLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			authModule.WriteError(w, r, authModule.Err.MissingRefreshToken, err)
			return
		default:
			authModule.WriteError(w, r, authModule.Err.ServerError, err)
			return
		}
	}

	err = authModule.Db.Tokens.DeleteByPlaintext(data.ScopeRefresh, cookie.Value)
	if err != nil {
		authModule.WriteError(w, r, authModule.Err.ServerError, err)
		return
	}

	newCookie := http.Cookie{
		Name:   "refresh_token",
		Path:   "/api/auth",
		MaxAge: -1,
	}

	http.SetCookie(w, &newCookie)
	http.Redirect(w, r, authModule.Cfg.BaseUrl, http.StatusFound)
}
