package feature

import (
	"context"
	"fmt"
	"user-service/config"
	"user-service/internal/application/domain/readmodel"
	"user-service/internal/ports/outbound"
)

type GetFollowingQry struct {
	Username string
	Cursor   int64
	PageSize int32
}

type GetFollowingRes struct {
	Following     []readmodel.UserSummary
	CurrentCursor int64
}

type GetFollowing struct {
	followRepo outbound.FollowRepository
}

func NewGetFollowing(followRepo outbound.FollowRepository) *GetFollowing {
	return &GetFollowing{
		followRepo: followRepo,
	}
}

func (gfu *GetFollowing) Execute(ctx context.Context, qry GetFollowingQry) (GetFollowingRes, error) {
	users, err := gfu.followRepo.GetFollowing(ctx, qry.Username, qry.PageSize, qry.Cursor/1000000.0)

	if err != nil {
		return GetFollowingRes{}, fmt.Errorf("GetFollowing.Execute: %w", err)
	}

	for _, u := range users {
		if u.AvatarThumbnailURL != nil {
			*u.AvatarThumbnailURL = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, *u.AvatarThumbnailURL)
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
