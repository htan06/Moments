package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/module/follow/domain"
)

type GetFollowingQry struct {
	Username string
	Page     int32
	PageSize int32
}

type GetFollowingUsecase struct {
	followRepo domain.FollowRepository
}

func NewGetFollowingUsecase(followRepo domain.FollowRepository) *GetFollowingUsecase {
	return &GetFollowingUsecase{
		followRepo: followRepo,
	}
}

func (gfu *GetFollowingUsecase) Execute(ctx context.Context, qry GetFollowingQry) ([]domain.UserSummary, error) {
	users, err := gfu.followRepo.GetFollowing(ctx, qry.Username, qry.PageSize, (qry.Page-1)*qry.PageSize)

	if err != nil {
		return nil, fmt.Errorf("GetFollowingUsecase.Execute: %w", err)
	}

	for _, u := range users {
		if u.AvatarThumbnailURL != nil {
			*u.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *u.AvatarThumbnailURL)
		}
	}
	return users, nil
}
