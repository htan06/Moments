package domain

import (
	"context"
	"io"
	"time"
)

type UserRepository interface {
	CreateProfile(ctx context.Context, p Profile) error
	GetAvatarIDByUserID(ctx context.Context, userID int64) (*string, error)
	UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) error
	GetSelfProfileByUsername(ctx context.Context, username string) (ProfileReadModel, error)
	GetOtherProfileByUsername(ctx context.Context, currentUserID int64, targetUsername string) (ProfileReadModel, error)
	UpdateAvatarIDAndAvatarThumbnailID(ctx context.Context, userID int64, avatarID string, avatarThumbnailID string) error
	FindProfilesByUsername(ctx context.Context, username string) ([]ProfileSummaryReadModel, error)
}

type CacheReposiotry interface {
	SetUploadAvatarSession(ctx context.Context, key string, avatarName string, ttl time.Duration) error
	GetUploadAvatarSession(ctx context.Context, key string) (string, error)
	RemoveSession(ctx context.Context, key string) error
}

type ObjectStorage interface {
	GetPresignedUrlUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error)
	Copy(ctx context.Context, bucketSrc string, objSrc string, bucketDest string, objDest string) error
	GetObject(ctx context.Context, bucket string, key string) (io.Reader, error)
	Upload(ctx context.Context, bucket string, key string, reader io.Reader, contentType string, size int64) error
	Remove(ctx context.Context, bucket string, key string) error
}

type ProcessImg interface {
	Resize(reader io.Reader, width int, height int) (io.Reader, int, error)
}
