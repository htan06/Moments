package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	GetAvatarIDByUserID(ctx context.Context, userID int64) (string, error)
	UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) error
	GetProfileByUsername(ctx context.Context, username string) (ProfileQry, error)
	UpdateAvatarID(ctx context.Context, userID int64, avatarID string) error
}

type CacheReposiotry interface {
	SetUploadAvatarSession(ctx context.Context, key string, avatarName string, ttl time.Duration) error
	GetUploadAvatarSession(ctx context.Context, key string) (string, error)
	RemoveSession(ctx context.Context, key string) error
}

type ObjectStorage interface {
	GetPresignedUrlUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error)
	PromoteAvatar(ctx context.Context, objName string) error
}
