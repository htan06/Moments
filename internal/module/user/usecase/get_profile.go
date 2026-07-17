package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/config"
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

func (gpu *GetProfileUsecase) Excute(ctx context.Context, username string) (domain.ProfileQry, error) {
	profile, err := gpu.userRepo.GetProfileByUsername(ctx, username)

	if err != nil {
		return domain.ProfileQry{}, fmt.Errorf("ChangeAvatarUsecase.Excute: %w", err)
	}

	if profile.AvatarURL != nil {
		*profile.AvatarURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *profile.AvatarURL)
	}
	return profile, nil
}
