package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, user User) error
	UpdateLastLogin(ctx context.Context, user User) error
	UpdatePassword(ctx context.Context, user User) error
	UpdateStatusActiveIfExistsProfile(ctx context.Context, userID int64) error
	GetUserDetailByID(ctx context.Context, userID int64) (UserDetail, error)
}

type CacheRepository interface {
	GetUserPending(ctx context.Context, key string) (UserPending, error)
	SetUserPendingIfNotExists(ctx context.Context, key string, u UserPending, ttl time.Duration) error

	GetActiveToken(ctx context.Context, key string) (string, error)
	SetActiveToken(ctx context.Context, key string, token string, ttl time.Duration) error

	Remove(ctx context.Context, key string) error
}

type EmailOTPSender interface {
	SendOTP(ctx context.Context, email string, otp string) error
}
