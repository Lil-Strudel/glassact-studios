package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

func (m *AuthModule) login(user data.AuthUser, w http.ResponseWriter) error {
	var plaintext string
	var expiry time.Time

	if user.IsDealership() {
		refreshToken, err := m.Db.DealershipTokens.New(user.GetID(), 30*24*time.Hour, data.DealershipScopeRefresh)
		if err != nil {
			return err
		}
		plaintext = refreshToken.Plaintext
		expiry = refreshToken.Expiry
	} else {
		refreshToken, err := m.Db.InternalTokens.New(user.GetID(), 30*24*time.Hour, data.InternalScopeRefresh)
		if err != nil {
			return err
		}
		plaintext = refreshToken.Plaintext
		expiry = refreshToken.Expiry
	}

	secure := false
	if m.Cfg.Env == "production" {
		secure = true
	}

	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    plaintext,
		Path:     "/api/auth",
		Expires:  expiry,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)

	return nil
}

func (m *AuthModule) configGoogle() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     m.Cfg.Google.ClientID,
		ClientSecret: m.Cfg.Google.ClientSecret,
		RedirectURL:  m.Cfg.Google.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

type googleInfoResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Verified bool   `json:"verified_email"`
	Picture  string `json:"picture"`
}

func getGoogleUserInfo(token string) (*googleInfoResponse, error) {
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

	var data googleInfoResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (m *AuthModule) configMicrosoft() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     m.Cfg.Microsoft.ClientID,
		ClientSecret: m.Cfg.Microsoft.ClientSecret,
		RedirectURL:  m.Cfg.Microsoft.RedirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     microsoft.AzureADEndpoint(""),
	}
}

type microsoftInfoResponse struct {
	Sub     string `json:"sub"`
	Picture string `json:"picture"`
	Email   string `json:"email"`
}

func getMicrosoftUserInfo(token string) (*microsoftInfoResponse, error) {
	reqURL, err := url.Parse("https://graph.microsoft.com/oidc/userinfo")
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

	var data microsoftInfoResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (m *AuthModule) getUserFromProvider(email, provider, providerID string) (data.AuthUser, bool, error) {
	existingAccount, found, err := m.Db.DealershipAccounts.GetByProvider(provider, providerID)
	if err != nil {
		return nil, false, err
	}

	if found {
		existingUser, found, err := m.Db.DealershipUsers.GetByID(existingAccount.DealershipUserID)
		if err != nil {
			return nil, false, err
		}

		if !found || !existingUser.IsActive {
			return nil, false, nil
		}

		return existingUser, true, nil
	}

	dealershipUser, found, err := m.Db.DealershipUsers.GetByEmail(email)
	if err != nil {
		return nil, false, err
	}

	if found && dealershipUser.IsActive {
		newAccount := data.DealershipAccount{
			DealershipUserID:  dealershipUser.ID,
			Type:              "oidc",
			Provider:          provider,
			ProviderAccountID: providerID,
		}

		err = m.Db.DealershipAccounts.Insert(&newAccount)
		if err != nil {
			return nil, false, err
		}

		return dealershipUser, true, nil
	}

	internalAccount, found, err := m.Db.InternalAccounts.GetByProvider(provider, providerID)
	if err != nil {
		return nil, false, err
	}

	if found {
		internalUser, found, err := m.Db.InternalUsers.GetByID(internalAccount.InternalUserID)
		if err != nil {
			return nil, false, err
		}

		if !found || !internalUser.IsActive {
			return nil, false, nil
		}

		return internalUser, true, nil
	}

	internalUser, found, err := m.Db.InternalUsers.GetByEmail(email)
	if err != nil {
		return nil, false, err
	}

	if found && internalUser.IsActive {
		newAccount := data.InternalAccount{
			InternalUserID:    internalUser.ID,
			Type:              "oidc",
			Provider:          provider,
			ProviderAccountID: providerID,
		}

		err = m.Db.InternalAccounts.Insert(&newAccount)
		if err != nil {
			return nil, false, err
		}

		return internalUser, true, nil
	}

	return nil, false, nil
}

func (m *AuthModule) emailMagicLink(email, token string) error {
	u, err := url.Parse(m.Cfg.BaseURL)
	if err != nil {
		return err
	}

	u.Path = path.Join(u.Path, "api", "auth", "magic-link", "callback")

	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	from := mail.Address{Name: "GlassAct Studios", Address: "no-reply@glassactstudios.com"}
	to := mail.Address{Address: email}

	subject := "Sign in to Glassact Studios"

	plain, html := generateMagicLinkEmail(u.String())
	message := buildMessage(from, to, subject, plain, html)

	auth := smtp.PlainAuth("", m.Cfg.Smtp.Username, m.Cfg.Smtp.Password, m.Cfg.Smtp.Host)

	err = smtp.SendMail(m.Cfg.Smtp.Host+":"+strconv.Itoa(m.Cfg.Smtp.Port), auth, from.Address, []string{to.Address}, message)
	if err != nil {
		return err
	}

	return nil
}

func randString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func buildMessage(from, to mail.Address, subject, textBody, htmlBody string) []byte {
	msgID := fmt.Sprintf("<%s@glassactstudios.com>", randString(12))
	date := time.Now().Format(time.RFC1123Z)
	boundary := "alt-" + randString(12)

	headers := ""
	headers += fmt.Sprintf("From: %s\r\n", from.String())
	headers += fmt.Sprintf("To: %s\r\n", to.String())
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	headers += fmt.Sprintf("Date: %s\r\n", date)
	headers += fmt.Sprintf("Message-ID: %s\r\n", msgID)

	body := ""
	body += fmt.Sprintf("--%s\r\n", boundary)
	body += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	body += "Content-Transfer-Encoding: 7bit\r\n"
	body += "\r\n"
	body += textBody + "\r\n"

	if htmlBody != "" {
		body += fmt.Sprintf("--%s\r\n", boundary)
		body += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
		body += "Content-Transfer-Encoding: 7bit\r\n"
		body += "\r\n"
		body += htmlBody + "\r\n"
	}

	body += fmt.Sprintf("--%s--\r\n", boundary)

	return []byte(headers + "\r\n" + body)
}

func generateMagicLinkEmail(magicLink string) (string, string) {
	textBody := fmt.Sprintf(`Sign in to Glassact Studios

Click the link below to securely sign in:

%s

If you did not request this, you can ignore this email.`, magicLink)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Glassact Studios – Sign In</title>
  </head>
  <body style="margin:0; padding:0; background-color:#ffffff; font-family:Roboto, Arial, sans-serif; color:#0a0a0a;">
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
      <tr>
        <td align="center" style="padding: 40px 0;">
          <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="max-width:600px; background:#ffffff; border-radius:8px; box-shadow:0 2px 4px rgba(0,0,0,0.1); padding:40px;">
            <tr>
              <td style="text-align:center;">
                <h1 style="margin:0; font-size:24px; font-weight:600; color:#0a0a0a;">Sign in to Glassact Studios</h1>
                <p style="margin:20px 0; font-size:16px; color:#737373;">Click the button below to securely sign in:</p>
                <a href="%s" style="display:inline-block; padding:12px 24px; background-color:#8b0f24; color:#ffffff; text-decoration:none; border-radius:8px; font-size:16px; font-weight:500;">
                  Sign In
                </a>
                <p style="margin-top:30px; font-size:14px; color:#737373;">
                  If you did not request this email, you can safely ignore it.
                </p>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`, magicLink)

	return textBody, htmlBody
}

func (m *AuthModule) generateSecureState() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	timestamp := time.Now().Unix()

	data := fmt.Sprintf("%s:%d", hex.EncodeToString(randomBytes), timestamp)

	mac := hmac.New(sha256.New, []byte(m.Cfg.AuthSecret))
	mac.Write([]byte(data))
	signature := mac.Sum(nil)

	stateData := fmt.Sprintf("%s:%s", data, hex.EncodeToString(signature))
	return base64.URLEncoding.EncodeToString([]byte(stateData)), nil
}

func (m *AuthModule) validateState(state string) error {
	if state == "" {
		return errors.New("missing state parameter")
	}

	decoded, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return errors.New("invalid state format")
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 3 {
		return errors.New("invalid state structure")
	}

	randomData, timestampStr, providedSig := parts[0], parts[1], parts[2]

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return errors.New("invalid timestamp in state")
	}

	if time.Now().Unix()-timestamp > int64(15*time.Minute.Seconds()) {
		return errors.New("state parameter has expired")
	}

	data := fmt.Sprintf("%s:%s", randomData, timestampStr)
	mac := hmac.New(sha256.New, []byte(m.Cfg.AuthSecret))
	mac.Write([]byte(data))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return errors.New("invalid state signature")
	}

	return nil
}
