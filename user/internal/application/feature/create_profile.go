package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/user/internal/domain"
)

type CreateProfileCmd struct {
	UserID   int64
	Name     string
	Username string
}

type CreateProfile struct {
	profileCreator ProfileCreator
}

func NewCreateProfile(
	profileCreator ProfileCreator,
) *CreateProfile {
	return &CreateProfile{
		profileCreator: profileCreator,
	}
}

func (uc *CreateProfile) Execute(ctx context.Context, cmd CreateProfileCmd) (*int64, error) {
	newProfile, err := domain.NewProfile(cmd.UserID, cmd.Name, cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("CreateProfile.Execute: %w", err)
	}

	if err := uc.profileCreator.CreateProfile(ctx, *newProfile); err != nil {
		return nil, fmt.Errorf("CreateProfile.Execute: %w", err)
	}
	return &cmd.UserID, nil
}
