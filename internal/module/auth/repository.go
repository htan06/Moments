package auth

import (
	"context"
	"time"

	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/model"
)

type UserReposiotry interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByID(ctx context.Context, id int64) (model.User, error)
	Create(ctx context.Context, user model.User) error
}

type CacheRepository interface {
	Get(ctx context.Context, key string) (interface{}, error)
	SetIfNotExists(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Remove(ctx context.Context, key string) error
}

type EmailOTPSender interface {
	Send(ctx context.Context, email string, otp string) error
}
