package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/domain"
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
		return domain.ProfileReadModel{}, errs.NewError(errs.NotFound, nil, domain.UserNotFound)
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
	return p, nil
}
