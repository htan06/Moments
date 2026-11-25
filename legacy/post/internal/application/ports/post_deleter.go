package ports

import (
	"context"
)

type PostDeleter interface {
	DeletePostByUserIDAndPostID(ctx context.Context, userID int64, postID int64) error
}

type PostLikeRemover interface {
	DeleteLikePost(ctx context.Context, userID int64, postID int64) error
}
