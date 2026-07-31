package domain

import (
	"context"
	"io"
	"time"
)

type UserRepository interface {
	GetIDByUsername(ctx context.Context, username string) (*int64, error)
}

type PostRepository interface {
	CreatePost(ctx context.Context, post Post) (*int64, error)
	GetPost(ctx context.Context, postID int64) (PostReadModel, error)
	GetPostsByUsername(ctx context.Context, username string) ([]PostSummary, error)
	DeletePostByUserIDAndPostID(ctx context.Context, userID int64, postID int64) error
}

type ObjectStorage interface {
	GetPresignedURLUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error)
	GetObject(ctx context.Context, bucketName string, objName string) (io.Reader, error)
	PromotePostImage(ctx context.Context, objName string) error
	PutObject(ctx context.Context, bucketName string, objName string, file io.Reader) error
}

type CacheRepository interface {
	SetUploadPostSession(ctx context.Context, key string, uploadPostSession UploadPostSession) error
	GetUploadPostSession(ctx context.Context, key string) (UploadPostSession, error)
}
