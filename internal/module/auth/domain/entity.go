package domain

import (
	"regexp"
	"time"

	"github.com/htan06/Moments/internal/errs"
)

type UserStatus string

var (
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneNumberRegex = regexp.MustCompile(`^(?:([+]\d{1,4})[-.\s]?)?(?:[(](\d{1,3})[)][-.\s]?)?(\d{1,4})[-.\s]?(\d{1,4})[-.\s]?(\d{1,9})$`)
)

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusPending  UserStatus = "PENDING"
)

type User struct {
	ID           int64
	Email        string
	PhoneNumber  *string
	PasswordHash string
	LastLoginAt  time.Time
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(
	email string,
	phoneNumber *string,
	passwordHash string,
) (*User, error) {

	// if !usernameRegex.MatchString(username) {
	// 	return nil, errs.NewError(errs.Invalid, nil, UsernameInvalid)
	// }

	// if len(name) == 0 || len(name) > 100 {
	// 	return nil, errs.NewError(errs.Invalid, nil, NameInvalid)
	// }

	if !emailRegex.MatchString(email) {
		return nil, errs.NewError(errs.Invalid, nil, EmailInvalid)
	}

	if phoneNumber != nil && *phoneNumber != "" {
		if !phoneNumberRegex.MatchString(*phoneNumber) {
			return nil, errs.NewError(errs.Invalid, nil, PhoneNumberInvalid)
		}
	} else {
		phoneNumber = nil
	}

	if passwordHash == "" {
		return nil, errs.NewError(errs.Invalid, nil, PasswordHashEmpty)
	}

	now := time.Now().UTC()

	return &User{
		// Username:     username,
		// Name:         name,
		Email:        email,
		PhoneNumber:  phoneNumber,
		PasswordHash: passwordHash,
		Status:       UserStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

type UserPending struct {
	OTP          string `json:"otp"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}
