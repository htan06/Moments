package friend

import (
	"context"

	"github.com/htan06/echo-messenger-rest-api/internal/module/friend/model"
)

type FriendRepository interface {
	FindByUsername(ctx context.Context, username string) (model.UserProfile, error)
	CreateFriendRequest(ctx context.Context, friendRequqest model.FriendRequest) error
}