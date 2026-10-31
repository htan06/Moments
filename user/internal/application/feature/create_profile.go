package feature

import (
	"context"
	"fmt"
	"user-service/internal/application/domain/entity"
	"user-service/internal/ports/outbound"
)

type CreateProfileCmd struct {
	UserID   int64
	Name     string
	Username string
}

type CreateProfile struct {
	userRepo outbound.UserRepository
}

func NewCreateProfile(
	userRepo outbound.UserRepository,
) *CreateProfile {
	return &CreateProfile{
		userRepo: userRepo,
	}
}

func (uc *CreateProfile) Execute(ctx context.Context, cmd CreateProfileCmd) (*int64, error) {
	newProfile, err := entity.NewProfile(cmd.UserID, cmd.Name, cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("CreateProfile.Execute: %w", err)
	}

	if err := uc.userRepo.CreateProfile(ctx, *newProfile); err != nil {
		return nil, fmt.Errorf("CreateProfile.Execute: %w", err)
	}
	return &cmd.UserID, nil
}
