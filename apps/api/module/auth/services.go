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
	var account model.Account

	err := database.Db.QueryRow(context.Background(), `
		SELECT *
		FROM accounts
		WHERE provider=$1 AND provider_account_id=$2
	`, provider, providerAccountId).Scan(&account)
	if err != nil {
		return nil, false
	}

	return &account, true
}

func GetUserByEmail(email string) (*model.User, bool) {
	rows, err := database.Db.Query(context.Background(), `
		SELECT *
		FROM users
		WHERE email=$1 
	`, email)
	if err != nil {
		fmt.Println(err)
		return nil, false
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		fmt.Println(err)
		return nil, false
	}

	return &user, true
}

func CreateNewUser(user model.User, account model.Account) error {
	tx, err := database.Db.Begin(context.Background())
	if err != nil {
		return err
	}

	defer tx.Rollback(context.Background())

	var userID int
	err = tx.QueryRow(context.Background(), `
		INSERT INTO users (email, name, email_verified, image) 
		VALUES (@Email, @Name, @EmailVerified, @Image)
		RETURNING id
	`, pgx.NamedArgs{
		"Email":         user.Email,
		"Name":          user.Name,
		"EmailVerified": user.EmailVerified,
		"Image":         user.Image,
	}).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(context.Background(), `
		INSERT INTO accounts (user_id, type, provider, provider_account_id, refresh_token, access_token, expires_at, id_token, scope, session_state, token_type)
		VALUES (@UserId, @Type, @Provider, @ProviderAccountId, @RefreshToken, @AccessToken, @ExpiresAt, @IdToken, @Scope, @SessionState, @TokenType)
	`, pgx.NamedArgs{
		"UserId":            userID,
		"Type":              "test",
		"Provider":          "asdf",
		"ProviderAccountId": "test",
		// "RefreshToken":      "",
		// "AccessToken":       "",
		// "ExpiresAt":         "",
		// "IdToken":           "",
		// "Scope":             "",
		// "SessionState":      "",
		// "TokenType":         "",
	})
	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}
