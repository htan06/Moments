// Package command holds the account module's write-side use cases: one file and
// one handler per use case. Dependencies are small interfaces declared here
// and implemented by the adapters.
package command

import (
	"context"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
	"github.com/htan06/Moments/application/internal/platform/token"
)

// Ports are split by capability (Reader, Creator, Updater, Deleter...) so each
// handler asks only for what it uses. When a handler needs several of them
// from the same object, a composite interface embeds them (SessionUpdateDeleter).

// AccountReader loads accounts. ByID / ByEmail / ByUsername return
// domain.ErrAccountNotFound when absent.
type AccountReader interface {
	EmailExists(ctx context.Context, email string) (bool, error)
	ByID(ctx context.Context, id int64) (domain.Account, error)
	ByEmail(ctx context.Context, email string) (domain.Account, error)
	ByUsername(ctx context.Context, username string) (domain.Account, error)
}

// AccountCreator stores a new account and returns its id. It returns
// domain.ErrEmailTaken / domain.ErrUsernameTaken on clashes (ACCOUNT-R7).
type AccountCreator interface {
	Create(ctx context.Context, a domain.Account) (int64, error)
}

// AccountDeleter removes an account; used to undo a registration whose
// profile could not be created (ACCOUNT-R7).
type AccountDeleter interface {
	Delete(ctx context.Context, id int64) error
}

// AccountCreateDeleter creates and removes accounts (one store does both).
type AccountCreateDeleter interface {
	AccountCreator
	AccountDeleter
}

// ProfileNameValidator checks a profile name with the user module, which owns
// Profile. It returns the normalized name or domain.ErrInvalidName.
type ProfileNameValidator interface {
	ValidateName(name string) (string, error)
}

// ProfileCreator asks the user module to create the profile of a new account
// (ACCOUNT-R7), with a copy of the username. It is idempotent, so a retry is
// safe. It returns domain.ErrUsernameTaken when the copy clashes.
type ProfileCreator interface {
	CreateProfile(ctx context.Context, userID int64, username, name string) error
}

// Change tells the verification store what to do after an update function.
type Change int

const (
	KeepSession   Change = iota // leave the stored session untouched
	SaveSession                 // store the modified session (lifetime unchanged)
	DeleteSession               // remove the session
)

// SessionCreator stores a new OTP verification session until it expires.
type SessionCreator interface {
	Create(ctx context.Context, s domain.VerificationSession) error
}

// SessionUpdater loads the session, runs fn and applies its Change atomically
// with respect to concurrent updates of the same session. It returns fn's
// error after applying the change, or domain.ErrSessionNotFound.
type SessionUpdater interface {
	Update(ctx context.Context, id string, fn func(*domain.VerificationSession) (Change, error)) error
}

// SessionDeleter removes a session.
type SessionDeleter interface {
	Delete(ctx context.Context, id string) error
}

// SessionUpdateDeleter updates and removes sessions (one store does both).
type SessionUpdateDeleter interface {
	SessionUpdater
	SessionDeleter
}

// OTPSender hands an OTP email to the mail service (asynchronously).
type OTPSender interface {
	SendOTP(ctx context.Context, email, otp string, expiresAt time.Time) error
}

// PasswordHasher hashes a new password.
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// PasswordVerifier compares a password with a stored hash. CompareDummy costs
// the same as Matches and is used when there is no account, so an unknown
// identifier is not faster to reject (ACCOUNT-R1).
type PasswordVerifier interface {
	Matches(hash, password string) bool
	CompareDummy(password string)
}

// LoginSubject is what failed login attempts are counted against (ACCOUNT-R3):
// the account when the identifier matches one (email and username share the
// counter), otherwise the identifier itself.
type LoginSubject struct {
	AccountID  int64
	Identifier string
}

// LoginAttemptCounter counts failed logins within the configured window.
type LoginAttemptCounter interface {
	// Blocked returns how long the subject stays blocked, or 0.
	Blocked(ctx context.Context, s LoginSubject) (time.Duration, error)
	Fail(ctx context.Context, s LoginSubject) error
}

// TokenIssuer signs a new access + refresh token pair.
type TokenIssuer interface {
	Issue(c token.Claims) (token.Pair, error)
}

// RefreshTokenParser verifies a refresh token and returns its claims.
type RefreshTokenParser interface {
	ParseRefresh(s string) (token.Claims, error)
}

// TokenParseIssuer verifies a refresh token and issues a new pair (one token
// manager does both).
type TokenParseIssuer interface {
	RefreshTokenParser
	TokenIssuer
}

func claimsOf(a domain.Account) token.Claims {
	return token.Claims{UserID: a.ID, Username: a.Username, Status: a.Status}
}
