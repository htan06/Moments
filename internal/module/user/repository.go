package user

import (
	"context"

	"github.com/htan06/echo-messenger-rest-api/internal/module/user/model"
)

type UserRepository interface {
	GetInfo(ctx context.Context, id int64) (model.User, error)
	UpdateInfo(ctx context.Context, user model.User) error
	ChangeReadStatus(ctx context.Context, user model.User) error
	UpdateUsername(ctx context.Context, user model.User) error
	FindByUsername(ctx context.Context, username string) (model.UserProfile, error)
}