package model

import "time"

type Account struct {
	ID                int        `json:"id" db:"id"`
	UserID            int        `json:"user_id" db:"user_id"`
	Type              string     `json:"type" db:"type"`
	Provider          string     `json:"provider" db:"provider"`
	ProviderAccountID string     `json:"provider_account_id" db:"provider_account_id"`
	RefreshToken      *string    `json:"refresh_token,omitempty" db:"refresh_token"`
	AccessToken       *string    `json:"access_token,omitempty" db:"access_token"`
	Expires           *time.Time `json:"expires,omitempty" db:"expires"`
	IDToken           *string    `json:"id_token,omitempty" db:"id_token"`
	Scope             *string    `json:"scope,omitempty" db:"scope"`
	SessionState      *string    `json:"session_state,omitempty" db:"session_state"`
	TokenType         *string    `json:"token_type,omitempty" db:"token_type"`
}
