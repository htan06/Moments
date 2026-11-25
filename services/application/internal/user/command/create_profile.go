// Package command holds the user module's write-side use cases: one file and
// one handler per use case. Dependencies are small interfaces declared here
// and implemented by the adapters.
package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/htan06/Moments/application/internal/user/domain"
)

// CreateProfile creates the profile of a newly created account (ACCOUNT-R7).
type CreateProfile struct {
	UserID   int64
	Username string // copy of the account's username
	Name     string
}

// ProfileCreator stores a new profile. It is idempotent: a profile that
// already exists for the user is left unchanged and no error is returned.
type ProfileCreator interface {
	Create(ctx context.Context, p domain.Profile) error
}

type CreateProfileHandler struct {
	profiles ProfileCreator
}

func NewCreateProfileHandler(profiles ProfileCreator) *CreateProfileHandler {
	return &CreateProfileHandler{profiles: profiles}
}

// Handle returns domain.ErrInvalidName for an invalid name and
// domain.ErrUsernameTaken when another profile holds the username. Calling it again
// for the same user succeeds without changing anything, so the caller can
// retry safely.
func (h *CreateProfileHandler) Handle(ctx context.Context, cmd CreateProfile) error {
	p, err := domain.NewProfile(cmd.UserID, cmd.Username, cmd.Name)
	if err != nil {
		return err
	}
	if err := h.profiles.Create(ctx, p); err != nil {
		if errors.Is(err, domain.ErrUsernameTaken) {
			return err
		}
		return fmt.Errorf("create profile: %w", err)
	}
	return nil
}
