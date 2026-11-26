package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/application/internal/user/domain"
)

// UpdateUsername replaces the username copy of a profile after the account
// module renamed the account (ACCOUNT-R5).
type UpdateUsername struct {
	UserID   int64
	Username string // new username, validated and lowercased by the account module
}

// ProfileUsernameUpdater replaces the username copy. It returns
// domain.ErrProfileNotFound / domain.ErrUsernameTaken.
type ProfileUsernameUpdater interface {
	UpdateUsername(ctx context.Context, userID int64, username string) error
}

type UpdateUsernameHandler struct {
	profiles ProfileUsernameUpdater
}

func NewUpdateUsernameHandler(profiles ProfileUsernameUpdater) *UpdateUsernameHandler {
	return &UpdateUsernameHandler{profiles: profiles}
}

// Handle returns domain.ErrInvalidUsername for an empty username,
// domain.ErrUsernameTaken when another profile holds it and
// domain.ErrProfileNotFound when the user has no profile. Setting the username
// the profile already has succeeds (the UPDATE matches the row).
func (h *UpdateUsernameHandler) Handle(ctx context.Context, cmd UpdateUsername) error {
	if err := domain.ValidateUsernameCopy(cmd.Username); err != nil {
		return err
	}
	err := h.profiles.UpdateUsername(ctx, cmd.UserID, cmd.Username)
	switch {
	case errors.Is(err, domain.ErrUsernameTaken), errors.Is(err, domain.ErrProfileNotFound):
		return err
	case err != nil:
		return fmt.Errorf("update username: %w", err)
	}
	return nil
}
