package usecase

import (
	"context"
	"fmt"
	"user-service/config"
	"user-service/internal/domain"
)

type GetFollowersQry struct {
	Username string
	Cursor   int64
	PageSize int32
}

type GetFollowersRes struct {
	Followers     []domain.UserSummary
	CurrentCursor int64
}

type GetFollowersUsecase struct {
	followRepo domain.FollowRepository
}

func NewGetFollowersUsecase(followRepo domain.FollowRepository) *GetFollowersUsecase {
	return &GetFollowersUsecase{
		followRepo: followRepo,
	}
}

func (gfu *GetFollowersUsecase) Execute(ctx context.Context, qry GetFollowersQry) (GetFollowersRes, error) {
	users, err := gfu.followRepo.GetFollowers(ctx, qry.Username, qry.PageSize, qry.Cursor)

	if err != nil {
		return GetFollowersRes{}, fmt.Errorf("GetFollowersUsecase.Execute: %w", err)
	}

	for _, u := range users {
		if u.AvatarThumbnailURL != nil {
			*u.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *u.AvatarThumbnailURL)
		}
	}
	var cursor int64
	if len(users) != 0 {
		cursor = users[len(users)-1].CreatedAt.UnixMicro()
	} else {
		cursor = 0
	}

	return GetFollowersRes{
		Followers:     users,
		CurrentCursor: cursor,
	}, nil
}
