package domain

import (
	"net/mail"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// StatusActive is the only account status for now (kept for extension).
const StatusActive = "ACTIVE"

const (
	usernameMinLen = 3
	usernameMaxLen = 30
	passwordMinLen = 6
	passwordMaxLen = 64
	nameMaxLen     = 50
	emailMaxLen    = 254
)

// Account is the login identity of a user.
type Account struct {
	ID           int64
	Email        string
	Username     string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Active reports whether the account may use authenticated features (USER-R6).
func (a Account) Active() bool { return a.Status == StatusActive }

// Profile is the public profile created together with the Account (USER-R7).
type Profile struct {
	UserID int64
	Name   string
}

// NewAccount builds a new ACTIVE account and its profile. email, username and
// passwordHash must already be validated / hashed. nameSuffix is the 4 random
// [a-z0-9] characters of the default profile name.
func NewAccount(email, username, passwordHash, nameSuffix string) (Account, Profile, error) {
	if passwordHash == "" {
		return Account{}, Profile{}, ErrInvalidPassword
	}
	name, err := ParseName(username + nameSuffix)
	if err != nil {
		return Account{}, Profile{}, err
	}
	return Account{
			Email:        email,
			Username:     username,
			PasswordHash: passwordHash,
			Status:       StatusActive,
		}, Profile{
			Name: name,
		}, nil
}

// ParseEmail validates an email and returns it lowercased (emails are unique
// case-insensitively).
func ParseEmail(raw string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" || len(s) > emailMaxLen {
		return "", ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(s)
	// Reject display names and other forms: only a bare address is accepted.
	if err != nil || addr.Address != s || addr.Name != "" {
		return "", ErrInvalidEmail
	}
	return s, nil
}

// UsernameRules validates usernames (USER-R5). Reserved holds lowercase names.
type UsernameRules struct {
	Reserved map[string]struct{}
}

// NewUsernameRules builds rules from a list of reserved names (any case).
func NewUsernameRules(reserved []string) UsernameRules {
	m := make(map[string]struct{}, len(reserved))
	for _, r := range reserved {
		if r = strings.ToLower(strings.TrimSpace(r)); r != "" {
			m[r] = struct{}{}
		}
	}
	return UsernameRules{Reserved: m}
}

// Parse validates a username and returns it lowercased.
func (r UsernameRules) Parse(raw string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if len(s) < usernameMinLen || len(s) > usernameMaxLen {
		return "", ErrInvalidUsername
	}
	for i := range len(s) {
		c := s[i]
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '.' || c == '_'
		if !ok {
			return "", ErrInvalidUsername
		}
	}
	if s[0] == '.' || s[len(s)-1] == '.' || strings.Contains(s, "..") {
		return "", ErrInvalidUsername
	}
	if _, reserved := r.Reserved[s]; reserved {
		return "", ErrInvalidUsername
	}
	return s, nil
}

// ValidatePassword checks USER-R4: 6–64 characters, at least one letter and
// one digit.
func ValidatePassword(p string) error {
	n := utf8.RuneCountInString(p)
	if !utf8.ValidString(p) || n < passwordMinLen || n > passwordMaxLen {
		return ErrInvalidPassword
	}
	var letter, digit bool
	for _, c := range p {
		switch {
		case unicode.IsLetter(c):
			letter = true
		case unicode.IsDigit(c):
			digit = true
		}
	}
	if !letter || !digit {
		return ErrInvalidPassword
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

// IsEmailIdentifier reports whether a login identifier is an email rather than
// a username (usernames never contain '@', USER-R5).
func IsEmailIdentifier(identifier string) bool {
	return strings.Contains(identifier, "@")
}
