package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type AuthModule struct {
	*app.Application
}

func NewAuthModule(app *app.Application) *AuthModule {
	return &AuthModule{
		app,
	}
}

func (am *AuthModule) HandleGetGoogleAuth(w http.ResponseWriter, r *http.Request) {
	state, err := am.generateSecureState()
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	google := am.configGoogle()
	url := google.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusFound)
}

func (am *AuthModule) HandleGetGoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if err := am.validateState(state); err != nil {
		am.WriteError(w, r, am.Err.BadRequest, err)
		return
	}

	token, err := am.configGoogle().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	user, found, err := am.getUserFromProvider(userInfo.Email, "google", userInfo.ID)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	if !found {
		am.WriteError(w, r, am.Err.AccountNotFound, nil)
		return
	}

	am.login(user.ID, w)
	http.Redirect(w, r, am.Cfg.BaseUrl, http.StatusFound)
}

func (am *AuthModule) HandleGetMicrosoftAuth(w http.ResponseWriter, r *http.Request) {
	state, err := am.generateSecureState()
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	microsoft := am.configMicrosoft()
	url := microsoft.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusFound)
}

func (am *AuthModule) HandleGetMicrosoftAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if err := am.validateState(state); err != nil {
		am.WriteError(w, r, am.Err.BadRequest, err)
		return
	}

	token, err := am.configMicrosoft().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	userInfo, err := getMicrosoftUserInfo(token.AccessToken)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	user, found, err := am.getUserFromProvider(userInfo.Email, "microsoft", userInfo.Sub)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	if !found {
		am.WriteError(w, r, am.Err.AccountNotFound, nil)
		return
	}

	am.login(user.ID, w)
	http.Redirect(w, r, am.Cfg.BaseUrl, http.StatusFound)
}

func (am *AuthModule) HandlePostMagicLinkAuth(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email" validate:"required,email"`
	}

	err := am.ReadJSONBody(w, r, &body)
	if err != nil {
		am.WriteError(w, r, am.Err.BadRequest, err)
		return
	}

	user, found, err := am.Db.Users.GetByEmail(body.Email)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}
	if !found {
		am.WriteError(w, r, am.Err.AccountNotFound, nil)
		return
	}

	loginToken, err := am.Db.Tokens.New(user.ID, 2*time.Hour, data.ScopeLogin)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	err = am.emailMagicLink(body.Email, loginToken.Plaintext)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	am.WriteJSON(w, r, http.StatusNoContent, nil)
}

func (am *AuthModule) HandleGetMagicLinkCallback(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	token := qs.Get("token")
	if token == "" {
		am.WriteError(w, r, am.Err.BadRequest, errors.New("missing token in query"))
		return
	}

	user, found, err := am.Db.Users.GetForToken(data.ScopeLogin, token)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	if !found {
		am.WriteError(w, r, am.Err.AccountNotFound, nil)
		return
	}

	am.login(user.ID, w)
	http.Redirect(w, r, am.Cfg.BaseUrl, http.StatusFound)
}

func (am *AuthModule) HandlePostTokenAccess(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			am.WriteError(w, r, am.Err.MissingRefreshToken, err)
			return
		default:
			am.WriteError(w, r, am.Err.ServerError, err)
			return
		}
	}

	user, found, err := am.Db.Users.GetForToken(data.ScopeRefresh, cookie.Value)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	if !found {
		am.WriteError(w, r, am.Err.AccountNotFound, nil)
		return
	}

	accessToken, err := am.Db.Tokens.New(user.ID, 2*time.Hour, data.ScopeAccess)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	am.WriteJSON(w, r, http.StatusCreated, map[string]any{
		"access_token":     accessToken.Plaintext,
		"access_token_exp": accessToken.Expiry,
	})
}

func (am *AuthModule) HandleGetLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			am.WriteError(w, r, am.Err.MissingRefreshToken, err)
			return
		default:
			am.WriteError(w, r, am.Err.ServerError, err)
			return
		}
	}

	err = am.Db.Tokens.DeleteByPlaintext(data.ScopeRefresh, cookie.Value)
	if err != nil {
		am.WriteError(w, r, am.Err.ServerError, err)
		return
	}

	newCookie := http.Cookie{
		Name:   "refresh_token",
		Path:   "/api/auth",
		MaxAge: -1,
	}

	http.SetCookie(w, &newCookie)
	http.Redirect(w, r, am.Cfg.BaseUrl, http.StatusFound)
}
