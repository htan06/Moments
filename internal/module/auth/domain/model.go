package domain

import "time"

type UserStatus string

const (
	UserActive    UserStatus = "ACTIVE"
	UserNonActive UserStatus = "NON_ACTIVE"
)

type User struct {
	ID           int64
	Username     string
	Name         string
	Email        string
	PhoneNumber  *string
	PasswordHash string
	LastLoginAt  time.Time
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserPending struct {
	OTP          string `json:"otp"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Name         string `json:"name"`
	Username     string `json:"username"`
}
