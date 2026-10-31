package outbound

import (
	"context"
	"user-service/internal/application/domain/entity"
	"user-service/internal/application/domain/readmodel"
)

type FollowRepository interface {
	CreateFollow(ctx context.Context, follow entity.Follow) error
	GetFollowing(ctx context.Context, username string, limit int32, cursor int64) ([]readmodel.UserSummary, error)
	GetFollowers(ctx context.Context, username string, limit int32, cursor int64) ([]readmodel.UserSummary, error)
	RemoveByIDAndFollowerID(ctx context.Context, f entity.Follow) error
	RemoveByIDAndFollowingID(ctx context.Context, f entity.Follow) error
}
