package friend

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
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

func (fs *FriendService) CreateFriendRequest(ctx context.Context, currentUserID int64, receiverUserID int64) (model.FriendRequest, error) {
	if currentUserID == receiverUserID {
		return model.FriendRequest{}, errs.NewError(errs.Invalid, nil, errs.FriendRequestInvalid)
	}
	friendRequest := model.FriendRequest{
		SenderID:   currentUserID,
		ReceiverID: receiverUserID,
	}

	fr, err := fs.friendRepo.CreateFriendRequest(ctx, friendRequest)
	if err != nil {
		return model.FriendRequest{}, fmt.Errorf("FriendService.FriendRequest %w", err)
	}
	return fr, nil
}

func (fs *FriendService) GetSentFriendRequests(ctx context.Context, currentUserID int64) ([]model.FriendRequestProfile, error) {
	friendRequest, err := fs.friendRepo.GetSentFriendRequests(ctx, currentUserID)
	if err != nil {
		return []model.FriendRequestProfile{}, fmt.Errorf("FriendService.GetSentFriendRequests: %w", err)
	}
	return friendRequest, nil
}

func (fs *FriendService) GetReceivedFriendRequests(ctx context.Context, currentUserID int64) ([]model.FriendRequestProfile, error) {
	friendRequest, err := fs.friendRepo.GetReceivedFriendRequests(ctx, currentUserID)
	if err != nil {
		return []model.FriendRequestProfile{}, fmt.Errorf("FriendService.GetPedingFriendRequests: %w", err)
	}
	return friendRequest, nil
}

func (fs *FriendService) GetListFriends(ctx context.Context, currentUserID int64) ([]model.UserProfile, error) {
	listFriends, err := fs.friendRepo.GetListFriends(ctx, currentUserID)
	if err != nil {
		return []model.UserProfile{}, fmt.Errorf("FriendService.GetListFriends: %w", err)
	}

	return listFriends, nil
}

func (fs *FriendService) AcceptFriendRequest(ctx context.Context, currentUserID int64, frID int64) error {

	if err := fs.friendRepo.AcceptFriendRequest(ctx, frID, currentUserID); err != nil {
		return fmt.Errorf("FriendService.AcceptFriendRequest: %w", err)
	}
	return nil
}

func (fs *FriendService) RejectFriendRequest(ctx context.Context, currentUserID int64, frID int64) error {

	if err := fs.friendRepo.RejectFriendRequest(ctx, frID, currentUserID); err != nil {
		return fmt.Errorf("FriendService.RejectFriendRequest: %w", err)
	}
	return nil
}

func (fs *FriendService) CancelFriendRequest(ctx context.Context, currentUserID int64, frID int64) error {

	if err := fs.friendRepo.CancelFriendRequest(ctx, frID, currentUserID); err != nil {
		return fmt.Errorf("FriendService.CancelFriendRequest: %w", err)
	}
	return nil
}

// func (fs *FriendService) UpdateStatusFriendRequest(ctx context.Context, requiestId int64, status model.FriendRequestStatus) error
// func (fs *FriendService) RemoveFriend(ctx context.Context, currentUserId int64, friendId int64) error
// func (fs *FriendService) GetBlockedUsers(ctx context.Context, currentUserId int64) ([]model.UserProfile, error)
// func (fs *FriendService) BlockUser(ctx context.Context, currentUserId int64, userBlockId int64) error
// func (fs *FriendService) UnblockUser(ctx context.Context, currentUserId int64, userUnblockId int64) error
