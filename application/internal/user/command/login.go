package command

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/htan06/Moments/application/internal/platform/token"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// Login signs in with email or username and password.
type Login struct {
	Identifier string
	Password   string
}

type LoginHandler struct {
	accounts  Accounts
	passwords Passwords
	attempts  LoginAttempts
	tokens    Tokens
	// dummyHash is compared when the identifier matches no account, so both
	// branches cost one hash comparison (USER-R1).
	dummyHash string
}

func NewLoginHandler(accounts Accounts, passwords Passwords, attempts LoginAttempts, tokens Tokens, dummyHash string) *LoginHandler {
	return &LoginHandler{accounts: accounts, passwords: passwords, attempts: attempts, tokens: tokens, dummyHash: dummyHash}
}

// Handle returns domain.ErrInvalidCredentials for unknown identifier and wrong
// password alike, and *domain.TooManyAttemptsError while blocked (USER-R3).
func (h *LoginHandler) Handle(ctx context.Context, cmd Login) (token.Pair, error) {
	identifier := strings.ToLower(strings.TrimSpace(cmd.Identifier))
	if identifier == "" || cmd.Password == "" {
		return token.Pair{}, domain.ErrInvalidCredentials
	}

	account, err := h.find(ctx, identifier)
	found := err == nil
	if err != nil && !errors.Is(err, domain.ErrAccountNotFound) {
		return token.Pair{}, fmt.Errorf("login: %w", err)
	}

	subject := LoginSubject{Identifier: identifier}
	if found {
		subject = LoginSubject{AccountID: account.ID}
	}
	wait, err := h.attempts.Blocked(ctx, subject)
	if err != nil {
		return token.Pair{}, fmt.Errorf("login: %w", err)
	}
	if wait > 0 {
		return token.Pair{}, &domain.TooManyAttemptsError{RetryAfterSeconds: int(math.Ceil(wait.Seconds()))}
	}

	hash := h.dummyHash
	if found {
		hash = account.PasswordHash
	}
	if !h.passwords.Matches(hash, cmd.Password) || !found {
		if err := h.attempts.Fail(ctx, subject); err != nil {
			return token.Pair{}, fmt.Errorf("login: %w", err)
		}
		return token.Pair{}, domain.ErrInvalidCredentials
	}
	// Success does not reset the failure counter (USER-R3).

	if !account.Active() {
		return token.Pair{}, domain.ErrAccountInactive
	}
	pair, err := h.tokens.Issue(claimsOf(account))
	if err != nil {
		return token.Pair{}, fmt.Errorf("login: %w", err)
	}
	return pair, nil
}

func (h *LoginHandler) find(ctx context.Context, identifier string) (domain.Account, error) {
	if domain.IsEmailIdentifier(identifier) {
		return h.accounts.ByEmail(ctx, identifier)
	}
	return h.accounts.ByUsername(ctx, identifier)
}
