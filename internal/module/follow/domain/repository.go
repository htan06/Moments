package domain

import (
	"context"
)

type FollowRepository interface {
	CreateFollow(ctx context.Context, follow Follow) (*int64, error)
	GetFollowing(ctx context.Context, username string, limit int32, offset int32) ([]UserSummary, error)
	GetFollowers(ctx context.Context, username string, limit int32, offset int32) ([]UserSummary, error)
	RemoveByFollowerID(ctx context.Context, followID int64, followerID int64) (Follow, error)
	RemoveByFollowingID(ctx context.Context, followID int64, followingID int64) (Follow, error)
}

type FollowProducer interface {
	SendMessage(ctx context.Context, followEvent FollowEvent) error
}
