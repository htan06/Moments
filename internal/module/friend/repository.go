package friend

import (
	"context"

	"github.com/htan06/echo-messenger-rest-api/internal/module/friend/model"
)

type FriendRepository interface {
	CreateFriendRequest(ctx context.Context, friendRequqest model.FriendRequest) (model.FriendRequest, error)
	GetSentFriendRequests(ctx context.Context, userID int64) ([]model.FriendRequestProfile, error)
	GetReceivedFriendRequests(ctx context.Context, userID int64) ([]model.FriendRequestProfile, error)
	CreateFriend(ctx context.Context, lowUserID int64, highUserID int64) error
	GetListFriends(ctx context.Context, userID int64) ([]model.UserProfile, error)

	AcceptFriendRequest(ctx context.Context, frID int64, userID int64) (error)
	RejectFriendRequest(ctx context.Context, frID int64, userID int64) error
	CancelFriendRequest(ctx context.Context, frID int64, userID int64) error
}