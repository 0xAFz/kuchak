package entity

import "time"

type User struct {
	ID              int       `json:"id"`
	Email           string    `json:"email"`
	Password        string    `json:"_"`
	IsEmailVerified bool      `json:"is_email_verified"`
	CreatedAt       time.Time `json:"created_at"`
}
