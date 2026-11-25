package ports

import (
	"context"

	"github.com/htan06/Moments/post/internal/domain/entity"
)

type PostCreator interface {
	CreatePost(ctx context.Context, post entity.Post) (*int64, error)
}

type PostLiker interface {
	CreateLikePost(ctx context.Context, userID int64, postID int64) error
}
