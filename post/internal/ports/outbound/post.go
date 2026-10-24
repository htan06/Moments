package outbound

import (
	"context"

	"github.com/htan06/Moments/post/internal/application/domain/entity"
	"github.com/htan06/Moments/post/internal/application/domain/readmodel"
)

type PostRepository interface {
	CreatePost(ctx context.Context, post entity.Post) (*int64, error)
	GetPost(ctx context.Context, postID int64) (readmodel.Post, error)
	GetPostsByAuthorID(ctx context.Context, viewerID int64, authorID int64, cursor int64, size int) ([]readmodel.PostSummary, error)
	DeletePostByUserIDAndPostID(ctx context.Context, userID int64, postID int64) error

	CreateLikePost(ctx context.Context, userID int64, postID int64) error
	DeleteLikePost(ctx context.Context, userID int64, postID int64) error

	GetRepostsByUsername(ctx context.Context, username string) ([]readmodel.PostSummary, error)
	CreateRepost(ctx context.Context, userID int64, postID int64) (int64, error)
	DeleteRepost(ctx context.Context, userID int64, repostID int64) error

	UpadateBatchLikeCount(ctx context.Context, list map[int64]int64) error
	GetLikeCount(ctx context.Context, postID int64) (int64, error)

	GetPostsByFollowing(ctx context.Context, viewerID int64, cursor int64, size int) ([]readmodel.Post, error)
	GetBatchPosts(ctx context.Context, viewerID int64, ids []int64) ([]readmodel.Post, error)
}
