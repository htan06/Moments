package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/module/follow/domain"
)

type GetFollowersQry struct {
	UserID   int64
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

func (gfu *GetFollowersUsecase) Excute(ctx context.Context, qry GetFollowersQry) ([]domain.UserSummary, error) {
	users, err := gfu.followRepo.GetFollowers(ctx, qry.UserID, qry.PageSize, qry.Page)

	if err != nil {
		return nil, fmt.Errorf("GetFollowersUsecase.Excute: %w", err)
	}

	for _, u := range users {
		if u.AvatarURL != nil {
			*u.AvatarURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *u.AvatarURL)
		}
	}
	return users, nil
}
