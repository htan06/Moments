package usecase

import (
	"context"
	"fmt"
	"user-service/config"
	"user-service/internal/domain"
	"user-service/internal/errs"
)

type GetProfileQry struct {
	CurrentUserID   int64
	CurrentUsername string
	TargetUsername  string
}

type GetProfileUsecase struct {
	userRepo domain.UserRepository
}

func NewGetProfileUsecase(userRepo domain.UserRepository) *GetProfileUsecase {
	return &GetProfileUsecase{
		userRepo: userRepo,
	}
}

func (gpu *GetProfileUsecase) Execute(ctx context.Context, qry GetProfileQry) (domain.ProfileReadModel, error) {
	if qry.TargetUsername == "" {
		return domain.ProfileReadModel{}, errs.NewError(errs.NotFound, nil, errs.UserNotFound)
	}

	var p domain.ProfileReadModel
	var err error

	if qry.CurrentUsername == qry.TargetUsername {
		p, err = gpu.userRepo.GetSelfProfileByUsername(ctx, qry.TargetUsername)
		p.IsOwner = true
	} else {
		p, err = gpu.userRepo.GetOtherProfileByUsername(ctx, qry.CurrentUserID, qry.TargetUsername)
		p.IsOwner = false
	}

	if err != nil {
		return domain.ProfileReadModel{}, fmt.Errorf("NewGetProfileUsecase.Execute: %w", err)
	}

	if p.AvatarURL != nil {
		*p.AvatarURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *p.AvatarURL)
	}

	if p.AvatarThumbnailURL != nil {
		*p.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *p.AvatarThumbnailURL)
	}

	return p, nil
}
