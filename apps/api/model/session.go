package model

import "time"

type Session struct {
	ID      int       `json:"id" db:"id"`
	UserID  int       `json:"user_id" db:"user_id"`
	Expires time.Time `json:"expires" db:"expires"`
	Data    []byte    `json:"data" db:"data"`
}
