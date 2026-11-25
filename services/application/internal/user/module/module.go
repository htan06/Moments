// Package module is the user module's facade: it wires the module from shared
// clients and exposes methods to other modules. Inputs and outputs are
// built-in types or types declared here, never domain entities.
package module

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/htan06/Moments/application/internal/user/adapter/postgres"
	"github.com/htan06/Moments/application/internal/user/command"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// Errors returned by the facade. Compare with errors.Is.
var (
	// ErrInvalidName: the profile name is rejected.
	ErrInvalidName = domain.ErrInvalidName
	// ErrUsernameTaken: another profile already holds the username copy.
	ErrUsernameTaken = domain.ErrUsernameTaken
)

type Deps struct {
	DB *pgxpool.Pool
}

type Module struct {
	createProfile *command.CreateProfileHandler
}

func New(d Deps) *Module {
	return &Module{
		createProfile: command.NewCreateProfileHandler(postgres.NewProfiles(d.DB)),
	}
}

// ValidateProfileName checks a profile name without storing anything and
// returns it normalized, or ErrInvalidName. The account module uses it to
// reject a bad name before the OTP is sent.
func (m *Module) ValidateProfileName(name string) (string, error) {
	return domain.ParseName(name)
}

// CreateProfile creates the profile of a new account (ACCOUNT-R7) and keeps a
// copy of its username, so reading a profile never asks the account module.
// It is idempotent: calling it again for the same user changes nothing. It
// returns ErrInvalidName for an invalid name and ErrUsernameTaken when another
// profile holds the username.
func (m *Module) CreateProfile(ctx context.Context, userID int64, username, name string) error {
	return m.createProfile.Handle(ctx, command.CreateProfile{UserID: userID, Username: username, Name: name})
}
