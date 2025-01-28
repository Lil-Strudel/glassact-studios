package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/Lil-Strudel/glassact-studios/apps/api/database"
	"github.com/Lil-Strudel/glassact-studios/apps/api/model"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func ConfigGoogle() *oauth2.Config {
	oauth := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
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

func GetGoogleUserInfo(token string) GoogleInfoResponse {
	reqURL, err := url.Parse("https://www.googleapis.com/oauth2/v1/userinfo")
	if err != nil {
		panic(err)
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
		panic(err)
	}

	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	var data GoogleInfoResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	return data
}

func FindExistingAccount(provider, providerAccountId string) (*model.Account, bool) {
	rows, err := database.Db.Query(context.Background(), `
		SELECT *
		FROM accounts
		WHERE provider=$1 AND provider_account_id=$2
	`, provider, providerAccountId)
	if err != nil {
		return nil, false
	}

	account, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Account])
	if err != nil {
		return nil, false
	}

	return &account, true
}

func GetUserByID(id int) (*model.User, bool) {
	rows, err := database.Db.Query(context.Background(), `
		SELECT *
		FROM users
		WHERE id=$1 
	`, id)
	if err != nil {
		return nil, false
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		return nil, false
	}

	return &user, true
}

func GetUserByEmail(email string) (*model.User, bool) {
	rows, err := database.Db.Query(context.Background(), `
		SELECT *
		FROM users
		WHERE email=$1 
	`, email)
	if err != nil {
		return nil, false
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		return nil, false
	}

	return &user, true
}

func CreateNewUser(user model.User, account model.Account) (*model.User, *model.Account, error) {
	tx, err := database.Db.Begin(context.Background())
	if err != nil {
		return nil, nil, err
	}

	defer tx.Rollback(context.Background())

	rows, err := tx.Query(context.Background(), `
		INSERT INTO users (email, name, email_verified, image) 
		VALUES (@Email, @Name, @EmailVerified, @Image)
		RETURNING *
	`, pgx.NamedArgs{
		"Email":         user.Email,
		"Name":          user.Name,
		"EmailVerified": user.EmailVerified,
		"Image":         user.Image,
	})
	if err != nil {
		return nil, nil, err
	}

	newUser, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		return nil, nil, err
	}

	rows, err = tx.Query(context.Background(), `
		INSERT INTO accounts (user_id, type, provider, provider_account_id, refresh_token, access_token, expires, id_token, scope, session_state, token_type)
		VALUES (@UserId, @Type, @Provider, @ProviderAccountId, @RefreshToken, @AccessToken, @Expires, @IdToken, @Scope, @SessionState, @TokenType)
		RETURNING *
	`, pgx.NamedArgs{
		"UserId":            newUser.ID,
		"Type":              account.Type,
		"Provider":          account.Provider,
		"ProviderAccountId": account.ProviderAccountID,
		"RefreshToken":      account.RefreshToken,
		"AccessToken":       account.AccessToken,
		"Expires":           account.Expires,
		"IdToken":           account.IDToken,
		"Scope":             account.Scope,
		"SessionState":      account.SessionState,
		"TokenType":         account.TokenType,
	})
	if err != nil {
		return nil, nil, err
	}

	newAccount, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Account])
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, nil, err
	}

	return &newUser, &newAccount, nil
}

func CreateNewAccount(account model.Account) (*model.Account, error) {
	rows, err := database.Db.Query(context.Background(), `
		INSERT INTO accounts (user_id, type, provider, provider_account_id, refresh_token, access_token, expires, id_token, scope, session_state, token_type)
		VALUES (@UserId, @Type, @Provider, @ProviderAccountId, @RefreshToken, @AccessToken, @Expires, @IdToken, @Scope, @SessionState, @TokenType)
		RETURNING *
	`, pgx.NamedArgs{
		"UserId":            account.UserID,
		"Type":              account.Type,
		"Provider":          account.Provider,
		"ProviderAccountId": account.ProviderAccountID,
		"RefreshToken":      account.RefreshToken,
		"AccessToken":       account.AccessToken,
		"Expires":           account.Expires,
		"IdToken":           account.IDToken,
		"Scope":             account.Scope,
		"SessionState":      account.SessionState,
		"TokenType":         account.TokenType,
	})
	if err != nil {
		return nil, err
	}

	acc, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Account])
	if err != nil {
		return nil, err
	}

	return &acc, nil
}
