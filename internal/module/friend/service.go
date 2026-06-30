package friend

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/module/friend/model"
)

type FriendService struct {
	friendRepo FriendRepository
}

func NewFriendService(
	friendRepo FriendRepository,
) *FriendService {
	return &FriendService{
		friendRepo: friendRepo,
	}
}

func (fs *FriendService) FindUserByUserName(ctx context.Context, username string) (model.UserProfile, error) {
	up, err := fs.friendRepo.FindByUsername(ctx, username)
	if err != nil {
		return model.UserProfile{}, fmt.Errorf("FriendService[FindUserByUserName]: %w", err)
	}
	return up, nil
}

func (fs *FriendService) CreateFriendRequest(ctx context.Context, currentUserID int64, receiverUserID int64) error {
	friendRequest := model.FriendRequest{
		SenderID: currentUserID,
		ReceiverID: receiverUserID,
		Status: model.FriendRequestStatusPending,
	}

	if err := fs.friendRepo.CreateFriendRequest(ctx, friendRequest); err != nil {
		return fmt.Errorf("FriendService.FriendRequest %w", err)
	}
	return nil
}
// func (fs *FriendService) UpdateStatusFriendRequest(ctx context.Context, requiestId int64, status model.FriendRequestStatus) error
// func (fs *FriendService) GetListFriends(ctx context.Context, currentUserId int64) ([]model.UserProfile, error)
// func (fs *FriendService) RemoveFriend(ctx context.Context, currentUserId int64, friendId int64) error
// func (fs *FriendService) GetBlockedUsers(ctx context.Context, currentUserId int64) ([]model.UserProfile, error)
// func (fs *FriendService) BlockUser(ctx context.Context, currentUserId int64, userBlockId int64) error
// func (fs *FriendService) UnblockUser(ctx context.Context, currentUserId int64, userUnblockId int64) error