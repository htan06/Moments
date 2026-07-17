package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	GetIDByUsername(ctx context.Context, username string) (int64, error)
}

type PostRepository interface {
	Create(ctx context.Context, post Post) (int64, error)
}

type ObjectStorage interface {
	GetPresignedURLUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error)
}
