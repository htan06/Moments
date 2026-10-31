package feature

import (
	"context"
	"fmt"
	"user-service/config"
	"user-service/internal/application/domain/readmodel"
	"user-service/internal/errs"
	"user-service/internal/ports/outbound"
)

type GetProfileQry struct {
	CurrentUserID   int64
	CurrentUsername string
	TargetUsername  string
}

type GetProfile struct {
	userRepo outbound.UserRepository
}

func NewGetProfile(userRepo outbound.UserRepository) *GetProfile {
	return &GetProfile{
		userRepo: userRepo,
	}
}

func (gpu *GetProfile) Execute(ctx context.Context, qry GetProfileQry) (readmodel.Profile, error) {
	if qry.TargetUsername == "" {
		return readmodel.Profile{}, errs.NewError(errs.NotFound, nil, errs.UserNotFound)
	}

	var p readmodel.Profile
	var err error

	if qry.CurrentUsername == qry.TargetUsername {
		p, err = gpu.userRepo.GetSelfProfileByUsername(ctx, qry.TargetUsername)
		p.IsOwner = true
	} else {
		p, err = gpu.userRepo.GetOtherProfileByUsername(ctx, qry.CurrentUserID, qry.TargetUsername)
		p.IsOwner = false
	}

	if err != nil {
		return readmodel.Profile{}, fmt.Errorf("NewGetProfile.Execute: %w", err)
	}

	if p.AvatarURL != nil {
		*p.AvatarURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, *p.AvatarURL)
	}

	if p.AvatarThumbnailURL != nil {
		*p.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, *p.AvatarThumbnailURL)
	}

	return p, nil
}
