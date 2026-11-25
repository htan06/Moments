// Package user bridges the account module to the user module's facade. It is
// the only place in the account module that imports another module.
package user

import (
	"context"
	"errors"

	"github.com/htan06/Moments/application/internal/account/domain"
	usermodule "github.com/htan06/Moments/application/internal/user/module"
)

// Profiles implements command.ProfileNameValidator and command.ProfileCreator
// on top of the user module, translating its errors into account errors.
type Profiles struct {
	users *usermodule.Module
}

func NewProfiles(users *usermodule.Module) *Profiles {
	return &Profiles{users: users}
}

func (p *Profiles) ValidateName(name string) (string, error) {
	n, err := p.users.ValidateProfileName(name)
	return n, translate(err)
}

func (p *Profiles) CreateProfile(ctx context.Context, userID int64, username, name string) error {
	return translate(p.users.CreateProfile(ctx, userID, username, name))
}

func translate(err error) error {
	switch {
	case errors.Is(err, usermodule.ErrInvalidName):
		return domain.ErrInvalidName
	case errors.Is(err, usermodule.ErrUsernameTaken):
		return domain.ErrUsernameTaken
	}
	return err
}
