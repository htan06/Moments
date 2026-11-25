package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/user/config"
	"github.com/htan06/Moments/user/internal/application/readmodel"
)

type GetFollowersQry struct {
	Username string
	Cursor   int64
	PageSize int32
}

type GetFollowersRes struct {
	Followers     []readmodel.UserSummary
	CurrentCursor int64
}

type GetFollowers struct {
	followReader FollowReader
}

func NewGetFollowers(followReader FollowReader) *GetFollowers {
	return &GetFollowers{
		followReader: followReader,
	}
}

func (gfu *GetFollowers) Execute(ctx context.Context, qry GetFollowersQry) (GetFollowersRes, error) {
	users, err := gfu.followReader.GetFollowers(ctx, qry.Username, qry.PageSize, qry.Cursor/1000000.0)

	if err != nil {
		return GetFollowersRes{}, fmt.Errorf("GetFollowers.Execute: %w", err)
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

	return GetFollowersRes{
		Followers:     users,
		CurrentCursor: cursor,
	}, nil
}
