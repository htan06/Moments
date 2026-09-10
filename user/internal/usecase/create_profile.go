package usecase

import (
	"context"
	"fmt"
	"user-service/internal/domain"
)

type CreateProfileCmd struct {
	UserID   int64
	Name     string
	Username string
}

type CreateProfileUC struct {
	userRepo domain.UserRepository
}

func NewCreateProfileUC(
	userRepo domain.UserRepository,
) *CreateProfileUC {
	return &CreateProfileUC{
		userRepo: userRepo,
	}
}

func (uc *CreateProfileUC) Execute(ctx context.Context, cmd CreateProfileCmd) (*int64, error) {
	newProfile, err := domain.NewProfile(cmd.UserID, cmd.Name, cmd.Username)
	if err != nil {
		return nil, fmt.Errorf("CreateProfileUC.Execute: %w", err)
	}

	if err := uc.userRepo.CreateProfile(ctx, *newProfile); err != nil {
		return nil, fmt.Errorf("CreateProfileUC.Execute: %w", err)
	}
	return &cmd.UserID, nil
}
