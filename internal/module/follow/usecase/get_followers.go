package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/module/follow/domain"
)

type GetFollowersQry struct {
	Username string
	Page     int32
	PageSize int32
}

type GetFollowersUsecase struct {
	followRepo domain.FollowRepository
}

func NewGetFollowersUsecase(followRepo domain.FollowRepository) *GetFollowersUsecase {
	return &GetFollowersUsecase{
		followRepo: followRepo,
	}
}

func (gfu *GetFollowersUsecase) Execute(ctx context.Context, qry GetFollowersQry) ([]domain.UserSummary, error) {
	users, err := gfu.followRepo.GetFollowers(ctx, qry.Username, qry.PageSize, (qry.Page-1)*qry.PageSize)

	if err != nil {
		return nil, fmt.Errorf("GetFollowersUsecase.Execute: %w", err)
	}

	for _, u := range users {
		if u.AvatarThumbnailURL != nil {
			*u.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *u.AvatarThumbnailURL)
		}
	}
	return users, nil
}
