package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/htan06/Moments/application/internal/platform/token"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// CompleteRegister is the last step: choose username and password on a
// verified session; creates Account + Profile and signs the user in.
type CompleteRegister struct {
	SessionID string
	Username  string
	Password  string
}

type CompleteRegisterHandler struct {
	accounts  Accounts
	sessions  VerificationSessions
	passwords Passwords
	tokens    Tokens
	random    Random
	usernames domain.UsernameRules
	now       func() time.Time
}

func NewCompleteRegisterHandler(accounts Accounts, sessions VerificationSessions, passwords Passwords,
	tokens Tokens, random Random, usernames domain.UsernameRules) *CompleteRegisterHandler {
	return &CompleteRegisterHandler{
		accounts:  accounts,
		sessions:  sessions,
		passwords: passwords,
		tokens:    tokens,
		random:    random,
		usernames: usernames,
		now:       time.Now,
	}
}

func (h *CompleteRegisterHandler) Handle(ctx context.Context, cmd CompleteRegister) (token.Pair, error) {
	// Input errors leave the session untouched so the user can correct them.
	username, err := h.usernames.Parse(cmd.Username)
	if err != nil {
		return token.Pair{}, err
	}
	if err := domain.ValidatePassword(cmd.Password); err != nil {
		return token.Pair{}, err
	}

	s, err := h.sessions.Get(ctx, cmd.SessionID)
	if err != nil {
		return token.Pair{}, err
	}
	if err := checkSession(&s, domain.VerificationRegister, h.now()); err != nil {
		return token.Pair{}, err
	}
	if s.Status != domain.VerificationVerified {
		return token.Pair{}, domain.ErrSessionNotVerified
	}

	hash, err := h.passwords.Hash(cmd.Password)
	if err != nil {
		return token.Pair{}, fmt.Errorf("complete register: %w", err)
	}
	suffix, err := h.random.NameSuffix()
	if err != nil {
		return token.Pair{}, err
	}
	account, profile, err := domain.NewAccount(s.Email, username, hash, suffix)
	if err != nil {
		return token.Pair{}, err
	}

	id, err := h.accounts.Create(ctx, account, profile)
	switch {
	case errors.Is(err, domain.ErrUsernameTaken):
		// Keep the session: the user picks another username.
		return token.Pair{}, err
	case errors.Is(err, domain.ErrEmailTaken):
		// Another registration for this email won the race (USER-R7).
		h.deleteSession(ctx, s.ID)
		return token.Pair{}, domain.ErrRegistrationFailed
	case err != nil:
		return token.Pair{}, fmt.Errorf("complete register: %w", err)
	}
	account.ID = id

	// The account exists now; a leftover session only expires on its own.
	h.deleteSession(ctx, s.ID)

	pair, err := h.tokens.Issue(claimsOf(account))
	if err != nil {
		return token.Pair{}, fmt.Errorf("complete register: %w", err)
	}
	return pair, nil
}

// deleteSession is best effort: a session that survives cannot be completed
// again because the email is now taken, and it expires on its own.
func (h *CompleteRegisterHandler) deleteSession(ctx context.Context, id string) {
	_ = h.sessions.Delete(ctx, id)
}
