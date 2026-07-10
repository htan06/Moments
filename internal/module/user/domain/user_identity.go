package domain

import (
	"time"
)

type UserStatus string
type NotificationService string
type UserActionType string

const (
	UserLocked UserStatus = "NON_ACTIVE"
	UserActive UserStatus = "ACTIVE"
)

type User struct {
	ID            int64
	Username      string
	Email         string
	PhoneNumber   string
	FirstName     string
	LastName      *string
	AvatarURL     *string
	CoverPhotoURL *string
	ReadStatus    *bool
	Status        UserStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
