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

	err = m.login(user, w)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

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

	err = m.login(user, w)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

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

	dealershipUser, found, err := m.Db.DealershipUsers.GetByEmail(body.Email)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if found && dealershipUser.IsActive {
		loginToken, err := m.Db.DealershipTokens.New(dealershipUser.ID, 2*time.Hour, data.DealershipScopeLogin)
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
		return
	}

	internalUser, found, err := m.Db.InternalUsers.GetByEmail(body.Email)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if found && internalUser.IsActive {
		loginToken, err := m.Db.InternalTokens.New(internalUser.ID, 2*time.Hour, data.InternalScopeLogin)
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
		return
	}

	m.WriteError(w, r, m.Err.AccountNotFound, nil)
}

func (m *AuthModule) HandleGetMagicLinkCallback(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	token := qs.Get("token")
	if token == "" {
		m.WriteError(w, r, m.Err.BadRequest, errors.New("missing token in query"))
		return
	}

	dealershipUser, found, err := m.Db.DealershipUsers.GetForToken(data.DealershipScopeLogin, token)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if found && dealershipUser.IsActive {
		err = m.login(dealershipUser, w)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		http.Redirect(w, r, m.Cfg.BaseURL, http.StatusFound)
		return
	}

	internalUser, found, err := m.Db.InternalUsers.GetForToken(data.InternalScopeLogin, token)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if found && internalUser.IsActive {
		err = m.login(internalUser, w)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		http.Redirect(w, r, m.Cfg.BaseURL, http.StatusFound)
		return
	}

	m.WriteError(w, r, m.Err.AccountNotFound, nil)
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

	user, _, err := data.GetAuthUserForToken(&m.Db, data.ScopeRefresh, cookie.Value)
	if err != nil {
		m.WriteError(w, r, m.Err.AccountNotFound, nil)
		return
	}

	if user == nil {
		m.WriteError(w, r, m.Err.AccountNotFound, nil)
		return
	}

	var accessToken *data.DealershipToken
	var internalAccessToken *data.InternalToken

	if user.IsDealership() {
		accessToken, err = m.Db.DealershipTokens.New(user.GetID(), 2*time.Hour, data.DealershipScopeAccess)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
	} else {
		internalAccessToken, err = m.Db.InternalTokens.New(user.GetID(), 2*time.Hour, data.InternalScopeAccess)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
	}

	var plaintext string
	var expiry time.Time

	if accessToken != nil {
		plaintext = accessToken.Plaintext
		expiry = accessToken.Expiry
	} else {
		plaintext = internalAccessToken.Plaintext
		expiry = internalAccessToken.Expiry
	}

	m.WriteJSON(w, r, http.StatusCreated, map[string]any{
		"access_token":     plaintext,
		"access_token_exp": expiry,
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

	err = m.Db.DealershipTokens.DeleteByPlaintext(data.DealershipScopeRefresh, cookie.Value)
	if err != nil {
		err = m.Db.InternalTokens.DeleteByPlaintext(data.InternalScopeRefresh, cookie.Value)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
	}

	newCookie := http.Cookie{
		Name:   "refresh_token",
		Path:   "/api/auth",
		MaxAge: -1,
	}

	http.SetCookie(w, &newCookie)
	http.Redirect(w, r, m.Cfg.BaseURL, http.StatusFound)
}
