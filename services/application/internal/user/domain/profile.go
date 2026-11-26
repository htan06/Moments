package domain

import (
	"strings"
	"unicode/utf8"
)

const nameMaxLen = 50

// Profile is the public profile of an account. UserID is the id of the
// Account owned by the account module (ACCOUNT-R7). Username is a copy of the
// account's username, kept here so reads do not ask the account module; the
// account module updates it when the username changes.
type Profile struct {
	UserID   int64
	Username string
	Name     string
}

// NewProfile builds the profile of a new account. username comes from the
// account module, already validated and lowercased.
func NewProfile(userID int64, username, name string) (Profile, error) {
	if err := ValidateUsernameCopy(username); err != nil {
		return Profile{}, err
	}
	name, err := ParseName(name)
	if err != nil {
		return Profile{}, err
	}
	return Profile{UserID: userID, Username: username, Name: name}, nil
}

// ValidateUsernameCopy checks the username copy. Its format belongs to the
// account module (ACCOUNT-R5), which sends it validated and lowercased; here it
// only has to be non-empty.
func ValidateUsernameCopy(username string) error {
	if username == "" {
		return ErrInvalidUsername
	}
	return nil
}

// ParseName validates a profile name: 1–50 characters after trimming.
func ParseName(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if n := utf8.RuneCountInString(s); n < 1 || n > nameMaxLen {
		return "", ErrInvalidName
	}
	return s, nil
}
