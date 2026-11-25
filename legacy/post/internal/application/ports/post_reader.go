package ports

import (
	"context"

	"github.com/htan06/Moments/post/internal/application/readmodel"
)

type PostByAuthorReader interface {
	GetPostsByAuthorID(ctx context.Context, viewerID int64, authorID int64, cursor int64, size int) ([]readmodel.Post, error)
}

type FollowingFeedReader interface {
	GetPostsByFollowing(ctx context.Context, viewerID int64, cursor int64, size int) ([]readmodel.Post, error)
}

type BatchPostReader interface {
	GetBatchPosts(ctx context.Context, viewerID int64, ids []int64) ([]readmodel.Post, error)
}
