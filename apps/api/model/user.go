package model

import "time"

type User struct {
	ID            int        `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	Name          *string    `json:"name,omitempty" db:"name"`
	EmailVerified *time.Time `json:"email_verified,omitempty" db:"email_verified"`
	Image         *string    `json:"image,omitempty" db:"image"`
}
