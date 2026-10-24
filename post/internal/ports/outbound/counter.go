package outbound

import "context"

type CounterRepository interface {
	IncPostLikes(ctx context.Context, postID int64) error
	SetPostLikesIfNotExists(ctx context.Context, postID int64, likeCount int) error
	DecPostLikes(ctx context.Context, postID int64) error
	GetLikeCount(ctx context.Context, postIDs ...int64) (map[int64]int64, error)
}
