package usecase

import (
	"context"
	"fmt"
	"user-service/config"
	"user-service/internal/domain"
)

type GetFollowingQry struct {
	Username string
	Cursor   int64
	PageSize int32
}

type GetFollowingRes struct {
	Following     []domain.UserSummary
	CurrentCursor int64
}

type GetFollowingUsecase struct {
	followRepo domain.FollowRepository
}

func NewGetFollowingUsecase(followRepo domain.FollowRepository) *GetFollowingUsecase {
	return &GetFollowingUsecase{
		followRepo: followRepo,
	}
}

func (gfu *GetFollowingUsecase) Execute(ctx context.Context, qry GetFollowingQry) (GetFollowingRes, error) {
	users, err := gfu.followRepo.GetFollowing(ctx, qry.Username, qry.PageSize, qry.Cursor)

	if err != nil {
		return GetFollowingRes{}, fmt.Errorf("GetFollowingUsecase.Execute: %w", err)
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

	return GetFollowingRes{
		Following:     users,
		CurrentCursor: cursor,
	}, nil
}
