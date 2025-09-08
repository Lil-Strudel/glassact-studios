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

func (m *AuthModule) HandleGetGoogleAuth(w http.ResponseWriter, r *http.Request) {
	state, err := m.generateSecureState()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	google := m.configGoogle()
	url := google.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusFound)
}

func (m *AuthModule) HandleGetGoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if err := m.validateState(state); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	token, err := m.configGoogle().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	userInfo, err := getGoogleUserInfo(token.AccessToken)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	user, found, err := m.getUserFromProvider(userInfo.Email, "google", userInfo.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.AccountNotFound, nil)
		return
	}

	m.login(user.ID, w)
	http.Redirect(w, r, m.Cfg.BaseURL, http.StatusFound)
}

func (m *AuthModule) HandleGetMicrosoftAuth(w http.ResponseWriter, r *http.Request) {
	state, err := m.generateSecureState()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	microsoft := m.configMicrosoft()
	url := microsoft.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusFound)
}

func (m *AuthModule) HandleGetMicrosoftAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if err := m.validateState(state); err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	token, err := m.configMicrosoft().Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	userInfo, err := getMicrosoftUserInfo(token.AccessToken)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	user, found, err := m.getUserFromProvider(userInfo.Email, "microsoft", userInfo.Sub)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.AccountNotFound, nil)
		return
	}

	m.login(user.ID, w)
	http.Redirect(w, r, m.Cfg.BaseURL, http.StatusFound)
}

func (m *AuthModule) HandlePostMagicLinkAuth(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email" validate:"required,email"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user, found, err := m.Db.Users.GetByEmail(body.Email)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.AccountNotFound, nil)
		return
	}

	loginToken, err := m.Db.Tokens.New(user.ID, 2*time.Hour, data.ScopeLogin)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	err = m.emailMagicLink(body.Email, loginToken.Plaintext)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusNoContent, nil)
}

func (m *AuthModule) HandleGetMagicLinkCallback(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	token := qs.Get("token")
	if token == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("missing token in query"))
		return
	}

	user, found, err := m.Db.Users.GetForToken(data.ScopeLogin, token)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.AccountNotFound, nil)
		return
	}

	m.login(user.ID, w)
	http.Redirect(w, r, m.Cfg.BaseURL, http.StatusFound)
}

func (m *AuthModule) HandlePostTokenAccess(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			m.WriteError(w, r, m.Err.MissingRefreshToken, err)
			return
		default:
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
	}

	user, found, err := m.Db.Users.GetForToken(data.ScopeRefresh, cookie.Value)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.AccountNotFound, nil)
		return
	}

	accessToken, err := m.Db.Tokens.New(user.ID, 2*time.Hour, data.ScopeAccess)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, map[string]any{
		"access_token":     accessToken.Plaintext,
		"access_token_exp": accessToken.Expiry,
	})
}

func (m *AuthModule) HandleGetLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			m.WriteError(w, r, m.Err.MissingRefreshToken, err)
			return
		default:
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
	}

	err = m.Db.Tokens.DeleteByPlaintext(data.ScopeRefresh, cookie.Value)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	newCookie := http.Cookie{
		Name:   "refresh_token",
		Path:   "/api/auth",
		MaxAge: -1,
	}

	http.SetCookie(w, &newCookie)
	http.Redirect(w, r, m.Cfg.BaseURL, http.StatusFound)
}
