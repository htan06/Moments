package command

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// ChangePassword changes the password of the signed-in user. UserID comes
// from the access token, never from the request body.
type ChangePassword struct {
	UserID      int64
	OldPassword string
	NewPassword string
}

type ChangePasswordHandler struct {
	accounts  AccountReadUpdater
	passwords PasswordHashVerifier
}

func NewChangePasswordHandler(accounts AccountReadUpdater, passwords PasswordHashVerifier) *ChangePasswordHandler {
	return &ChangePasswordHandler{accounts: accounts, passwords: passwords}
}

// Handle checks the current password, then that the new one differs from it
// (ACCOUNT-R4), by comparing the new password with the stored hash: only the
// hash is kept, so there is no other way to compare. Wrong current passwords
// are not limited (dev decision; see ACCOUNT-R3 in CLAUDE.md). Tokens already
// issued stay valid (ACCOUNT-R6: only username / status changes reissue).
//
// Errors: domain.ErrInvalidPassword (new password malformed),
// domain.ErrWrongPassword, domain.ErrSamePassword, domain.ErrAccountNotFound,
// domain.ErrAccountInactive.
func (h *ChangePasswordHandler) Handle(ctx context.Context, cmd ChangePassword) error {
	if err := domain.ValidatePassword(cmd.NewPassword); err != nil {
		return err
	}
	account, err := h.accounts.ByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	// The token may predate a status change (ACCOUNT-R6), so check again.
	if !account.Active() {
		return domain.ErrAccountInactive
	}
	if !h.passwords.Matches(account.PasswordHash, cmd.OldPassword) {
		return domain.ErrWrongPassword
	}
	if h.passwords.Matches(account.PasswordHash, cmd.NewPassword) {
		return domain.ErrSamePassword
	}
	hash, err := h.passwords.Hash(cmd.NewPassword)
	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	if err := account.SetPasswordHash(hash); err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	if err := h.accounts.UpdatePasswordHash(ctx, account.ID, account.PasswordHash); err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	return nil
}
