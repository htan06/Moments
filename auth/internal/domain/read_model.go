package domain

import "time"

type UserDetail struct {
	Username    string    `json:"username" db:"username"`
	Email       string    `json:"email" db:"email"`
	PhoneNumber *string   `json:"phone_number" db:"phone_number"`
	LastLoginAt   time.Time `json:"last_login_at" db:"last_login_at"`
}
