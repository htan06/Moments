package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// ResetPassword finishes the forgot-password flow. The otp is checked in this
// same request (ACCOUNT-R2): holding the session id alone is not enough.
type ResetPassword struct {
	SessionID   string
	OTP         string
	NewPassword string
}

type ResetPasswordHandler struct {
	accounts  AccountReadUpdater
	sessions  SessionUpdateDeleter
	passwords PasswordHasher
	policy    domain.OTPPolicy
	now       func() time.Time
}

func NewResetPasswordHandler(accounts AccountReadUpdater, sessions SessionUpdateDeleter, passwords PasswordHasher,
	policy domain.OTPPolicy) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		accounts:  accounts,
		sessions:  sessions,
		passwords: passwords,
		policy:    policy,
		now:       time.Now,
	}
}

// Handle sets the new password when the otp is right. The new password is
// validated first, so a malformed one does not use up an otp attempt. Wrong
// otps are counted; reaching the limit deletes the session (ACCOUNT-R2). A
// dummy session rejects every otp, like a wrong one.
//
// Tokens already issued stay valid until they expire (ACCOUNT-R6: logging out
// only happens on the client). The failed-login counter is not cleared yet
// (ACCOUNT-R3, waiting for the shared failed-attempt design).
func (h *ResetPasswordHandler) Handle(ctx context.Context, cmd ResetPassword) error {
	if err := domain.ValidatePassword(cmd.NewPassword); err != nil {
		return err
	}
	var s domain.VerificationSession
	err := h.sessions.Update(ctx, cmd.SessionID, func(cur *domain.VerificationSession) (Change, error) {
		if err := checkSession(cur, domain.VerificationForgotPassword, h.now()); err != nil {
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
		return err
	}

	account, err := h.accounts.ByEmail(ctx, s.Email)
	if errors.Is(err, domain.ErrAccountNotFound) {
		// Removed since the session started (only a failed registration is
		// ever undone): nothing left to reset.
		_ = h.sessions.Delete(ctx, s.ID)
		return domain.ErrSessionNotFound
	}
	if err != nil {
		// Keep the session: the user retries with the same otp.
		return fmt.Errorf("reset password: %w", err)
	}
	hash, err := h.passwords.Hash(cmd.NewPassword)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if err := account.SetPasswordHash(hash); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	if err := h.accounts.UpdatePasswordHash(ctx, account.ID, account.PasswordHash); err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	// Best effort: a surviving session expires on its own and still needs the
	// otp, which only the owner of the email has.
	_ = h.sessions.Delete(ctx, s.ID)
	return nil
}
