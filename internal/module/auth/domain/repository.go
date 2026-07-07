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
}

type CacheRepository interface {
	GetUserPending(ctx context.Context, key string) (UserPending, error)
	SetUserPendingIfNotExists(ctx context.Context, key string, u UserPending, ttl time.Duration) error
	RemoveUserPending(ctx context.Context, key string) error
}

type EmailOTPSender interface {
	SendOTP(ctx context.Context, email string, otp string) error
}
