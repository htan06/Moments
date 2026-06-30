package model

import "time"

type FriendRequestStatus string

const (
	FriendRequestStatusPending  FriendRequestStatus = "PENDING"
	FriendRequestStatusCanceled FriendRequestStatus = "CANCELED"
	FriendRequestStatusRejected FriendRequestStatus = "REJECTED"
	FriendRequestStatusAccepted FriendRequestStatus = "ACCEPTED"
)

type UserFriend struct {
	FriendID  int64
	CreatedAt time.Time
}

type FriendRequest struct {
	SenderID   int64
	ReceiverID int64
	Status     FriendRequestStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ReceivedFriendRequest struct {
	FromUserID int64
	Status     FriendRequestStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type BlockedUser struct {
	BlockUserID int64
	CreatedAt   time.Time
}
