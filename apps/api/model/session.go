package model

import "time"

type Session struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Expires      time.Time `json:"expires" db:"expires"`
	SessionToken string    `json:"session_token" db:"session_token"`
}
