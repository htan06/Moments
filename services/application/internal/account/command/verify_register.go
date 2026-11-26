package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
	"github.com/htan06/Moments/application/internal/platform/token"
)

// VerifyRegister checks the otp of a registration session and, when it is
// right, creates the account in the same request (ACCOUNT-R2): holding the
// session id alone is not enough to create an account.
type VerifyRegister struct {
	SessionID string
	OTP       string
}

type VerifyRegisterHandler struct {
	accounts AccountCreateDeleter
	profiles ProfileCreator
	sessions SessionUpdateDeleter
	tokens   TokenIssuer
	policy   domain.OTPPolicy
	now      func() time.Time
}

func NewVerifyRegisterHandler(accounts AccountCreateDeleter, profiles ProfileCreator, sessions SessionUpdateDeleter,
	tokens TokenIssuer, policy domain.OTPPolicy) *VerifyRegisterHandler {
	return &VerifyRegisterHandler{
		accounts: accounts,
		profiles: profiles,
		sessions: sessions,
		tokens:   tokens,
		policy:   policy,
		now:      time.Now,
	}
}

// Handle creates the Account, asks the user module to create its Profile and
// signs the user in. Wrong otps are counted; reaching the limit deletes the session (ACCOUNT-R2).
// The session is also deleted, and the user registers again, when the
// username was taken meanwhile or the email was registered by another flow
// (ACCOUNT-R7).
func (h *VerifyRegisterHandler) Handle(ctx context.Context, cmd VerifyRegister) (token.Pair, error) {
	var s domain.VerificationSession
	err := h.sessions.Update(ctx, cmd.SessionID, func(cur *domain.VerificationSession) (Change, error) {
		if err := checkSession(cur, domain.VerificationRegister, h.now()); err != nil {
			return KeepSession, err
		}
		res, err := cur.Verify(cmd.OTP, h.policy)
		switch res {
		case domain.VerifyDeleteWrong:
			return DeleteSession, err
		case domain.VerifyWrong:
			return SaveSession, err
		}
		s = *cur
		return KeepSession, nil
	})
	if err != nil {
		return token.Pair{}, err
	}

	reg := s.Registration
	account, err := domain.NewAccount(s.Email, reg.Username, reg.PasswordHash)
	if err != nil {
		// The data was validated when the session was created.
		return token.Pair{}, fmt.Errorf("verify register: invalid session data: %w", err)
	}
	id, err := h.accounts.Create(ctx, account)
	switch {
	case errors.Is(err, domain.ErrUsernameTaken):
		// Taken while waiting for the otp: the username was not reserved.
		h.deleteSession(ctx, s.ID)
		return token.Pair{}, err
	case errors.Is(err, domain.ErrEmailTaken):
		// Another registration for this email won the race (ACCOUNT-R7).
		h.deleteSession(ctx, s.ID)
		return token.Pair{}, domain.ErrRegistrationFailed
	case err != nil:
		// Keep the session: the user can retry with the same otp.
		return token.Pair{}, fmt.Errorf("verify register: %w", err)
	}
	account.ID = id

	// Account and Profile live in different modules, so they cannot share a
	// transaction. If the profile fails, undo the account (ACCOUNT-R7).
	if err := h.profiles.CreateProfile(ctx, id, reg.Username, reg.Name); err != nil {
		if derr := h.accounts.Delete(ctx, id); derr != nil {
			return token.Pair{}, fmt.Errorf("verify register: create profile: %w; undo account %d: %v", err, id, derr)
		}
		if errors.Is(err, domain.ErrUsernameTaken) {
			// The profile copy still holds the username (e.g. a rename saga
			// in progress): same outcome as a clash in the account table.
			h.deleteSession(ctx, s.ID)
			return token.Pair{}, err
		}
		// Keep the session: the user retries with the same otp.
		return token.Pair{}, fmt.Errorf("verify register: create profile: %w", err)
	}
	h.deleteSession(ctx, s.ID)

	pair, err := h.tokens.Issue(claimsOf(account))
	if err != nil {
		return token.Pair{}, fmt.Errorf("verify register: %w", err)
	}
	return pair, nil
}

// deleteSession is best effort: a session that survives cannot create the
// account again because the email is now taken, and it expires on its own.
func (h *VerifyRegisterHandler) deleteSession(ctx context.Context, id string) {
	_ = h.sessions.Delete(ctx, id)
}
