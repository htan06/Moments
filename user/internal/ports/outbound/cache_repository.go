package outbound

import (
	"context"
	"time"
)

type CacheReposiotry interface {
	SetUploadAvatarSession(ctx context.Context, key string, avatarName string, ttl time.Duration) error
	GetUploadAvatarSession(ctx context.Context, key string) (string, error)
	RemoveSession(ctx context.Context, key string) error
}
