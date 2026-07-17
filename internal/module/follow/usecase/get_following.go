package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/module/follow/domain"
)

type GetFollowingQry struct {
	UserID   int64
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

func (gfu *GetFollowingUsecase) Excute(ctx context.Context, qry GetFollowingQry) ([]domain.UserSummary, error) {
	users, err := gfu.followRepo.GetFollowing(ctx, qry.UserID, qry.PageSize, qry.Page)

	if err != nil {
		return nil, fmt.Errorf("GetFollowingUsecase.Excute: %w", err)
	}

	for _, u := range users {
		if u.AvatarURL != nil {
			*u.AvatarURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *u.AvatarURL)
		}
	}
	return users, nil
}
