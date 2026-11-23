// Package command holds the user module's write-side use cases: one file and
// one handler per use case. Dependencies are small interfaces declared here
// and implemented by the adapters.
package command

import (
	"context"
	"time"

	"github.com/htan06/Moments/application/internal/platform/token"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// Accounts persists Account + Profile.
type Accounts interface {
	EmailExists(ctx context.Context, email string) (bool, error)
	// Create stores account and profile in one transaction and returns the new
	// id. It returns domain.ErrEmailTaken / domain.ErrUsernameTaken on clashes.
	Create(ctx context.Context, a domain.Account, p domain.Profile) (int64, error)
	// ByID / ByEmail / ByUsername return domain.ErrAccountNotFound when absent.
	ByID(ctx context.Context, id int64) (domain.Account, error)
	ByEmail(ctx context.Context, email string) (domain.Account, error)
	ByUsername(ctx context.Context, username string) (domain.Account, error)
}

// Change tells the verification store what to do after an update function.
type Change int

const (
	KeepSession   Change = iota // leave the stored session untouched
	SaveSession                 // store the modified session (lifetime unchanged)
	DeleteSession               // remove the session
)

// VerificationSessions stores OTP verification sessions until they expire.
type VerificationSessions interface {
	Create(ctx context.Context, s domain.VerificationSession) error
	// Get returns domain.ErrSessionNotFound when absent or expired.
	Get(ctx context.Context, id string) (domain.VerificationSession, error)
	// Update loads the session, runs fn and applies its Change atomically with
	// respect to concurrent updates of the same session. It returns fn's error
	// after applying the change, or domain.ErrSessionNotFound.
	Update(ctx context.Context, id string, fn func(*domain.VerificationSession) (Change, error)) error
	Delete(ctx context.Context, id string) error
}

// OTPSender hands an OTP email to the mail service (asynchronously).
type OTPSender interface {
	SendOTP(ctx context.Context, email, otp string, expiresAt time.Time) error
}

// Passwords hashes and compares passwords.
type Passwords interface {
	Hash(password string) (string, error)
	Matches(hash, password string) bool
}

// LoginSubject is what failed login attempts are counted against (USER-R3):
// the account when the identifier matches one (email and username share the
// counter), otherwise the identifier itself.
type LoginSubject struct {
	AccountID  int64
	Identifier string
}

// LoginAttempts counts failed logins within the configured window.
type LoginAttempts interface {
	// Blocked returns how long the subject stays blocked, or 0.
	Blocked(ctx context.Context, s LoginSubject) (time.Duration, error)
	Fail(ctx context.Context, s LoginSubject) error
}

// Tokens issues and verifies JWTs.
type Tokens interface {
	Issue(c token.Claims) (token.Pair, error)
	ParseRefresh(s string) (token.Claims, error)
}

func claimsOf(a domain.Account) token.Claims {
	return token.Claims{UserID: a.ID, Username: a.Username, Status: a.Status}
}
