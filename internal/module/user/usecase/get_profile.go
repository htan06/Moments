package usecase

import (
	"context"

	"github.com/htan06/echo-messenger-rest-api/internal/module/user/domain"
)

type GetProfileUsecase struct {
	userRepo domain.UserRepository
}

func NewGetProfileUsecase(userRepo domain.UserRepository) *GetProfileUsecase {
	return &GetProfileUsecase{
		userRepo: userRepo,
	}
}

func (gpu *GetProfileUsecase) Excute(ctx context.Context, username string) (domain.UserProfile, error) {
	up, err := gpu.userRepo.GetProfileByUsername(ctx, username)
	if err != nil {
		return domain.UserProfile{}, err
	}
	return up, nil
}