package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// StatusActive is the only account status for now (kept for extension).
const StatusActive = "ACTIVE"

const (
	usernameMinLen = 3
	usernameMaxLen = 30
	passwordMinLen = 6
	passwordMaxLen = 64
	emailMaxLen    = 254
)

var (
	// emailPattern accepts a bare address: local part of [a-z0-9._%+-], a
	// domain of dot-separated labels that do not start or end with '-'.
	// Input is lowercased before matching.
	emailPattern = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`)
	// usernamePattern is ACCOUNT-R5 without the length: [a-z0-9_] groups joined by
	// single dots, so no leading / trailing dot and no "..".
	usernamePattern = regexp.MustCompile(`^[a-z0-9_]+(\.[a-z0-9_]+)*$`)
	// passwordLetter / passwordDigit: ACCOUNT-R4 needs at least one of each
	// (any script, e.g. "mậtkhẩu1").
	passwordLetter = regexp.MustCompile(`\pL`)
	passwordDigit  = regexp.MustCompile(`\p{Nd}`)
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

// Active reports whether the account may use authenticated features (ACCOUNT-R6).
func (a Account) Active() bool { return a.Status == StatusActive }

// NewAccount builds a new ACTIVE account. email, username and passwordHash
// must already be validated / hashed. Its profile belongs to the user module,
// which creates it right after the account (ACCOUNT-R7).
func NewAccount(email, username, passwordHash string) (Account, error) {
	if passwordHash == "" {
		return Account{}, ErrInvalidPassword
	}
	return Account{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		Status:       StatusActive,
	}, nil
}

// SetPasswordHash replaces the password with a new, already hashed one. The
// forgot-password flow does not require it to differ from the old password:
// the user does not know the old one, and telling them it matches would
// reveal it.
func (a *Account) SetPasswordHash(hash string) error {
	if hash == "" {
		return ErrInvalidPassword
	}
	a.PasswordHash = hash
	return nil
}

// ParseEmail validates an email and returns it lowercased (emails are unique
// case-insensitively).
func ParseEmail(raw string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(raw))
	if len(s) > emailMaxLen || !emailPattern.MatchString(s) {
		return "", ErrInvalidEmail
	}
	return s, nil
}

// UsernameRules validates usernames (ACCOUNT-R5). Reserved holds lowercase names.
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
	// The pattern only allows ASCII, so len counts characters.
	if len(s) < usernameMinLen || len(s) > usernameMaxLen || !usernamePattern.MatchString(s) {
		return "", ErrInvalidUsername
	}
	if _, reserved := r.Reserved[s]; reserved {
		return "", ErrInvalidUsername
	}
	return s, nil
}

// ValidatePassword checks ACCOUNT-R4: 6–64 characters, at least one letter and
// one digit.
func ValidatePassword(p string) error {
	n := utf8.RuneCountInString(p)
	if !utf8.ValidString(p) || n < passwordMinLen || n > passwordMaxLen {
		return ErrInvalidPassword
	}
	if !passwordLetter.MatchString(p) || !passwordDigit.MatchString(p) {
		return ErrInvalidPassword
	}
	return nil
}

// IsEmailIdentifier reports whether a login identifier is an email rather than
// a username (usernames never contain '@', ACCOUNT-R5).
func IsEmailIdentifier(identifier string) bool {
	return strings.Contains(identifier, "@")
}
