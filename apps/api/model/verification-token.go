package model

import "time"

type VerfificationToken struct {
	Identifier string    `json:"identifier" db:"identifier"`
	Expires    time.Time `json:"expires" db:"expires"`
	Token      string    `json:"token" db:"token"`
}
