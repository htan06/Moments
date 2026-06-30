package model

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string
type NotificationService string
type UserActionType string

const (
	UserLocked UserStatus = "LOCKED"
	UserActive UserStatus = "ACTIVE"

	FcmService NotificationService = "FCM"
	ApnService NotificationService = "APN"
	WebService NotificationService = "WEB"

	UserActionAddFriend           UserActionType = "ADD_FRIEND"
	UserActionRejectFriendRequest UserActionType = "REJECT_FRIEND_REQUEST"
	UserActionCancelFriendRequest UserActionType = "CANCEL_FRIEND_REQUEST"
	UserActionBlockUser           UserActionType = "BLOCK_USER"
	UserActionUnblockUser         UserActionType = "UNBLOCK_USER"
	UserActionLogin               UserActionType = "LOGIN"
	UserActionLogout              UserActionType = "LOGOUT"
)

type User struct {
	ID            int64
	Username      string
	Email         string
	PhoneNumber   string
	FirstName     string
	LastName      *string
	AvatarURL     *string
	CoverPhotoURL *string
	ReadStatus    *bool
	Status        UserStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// UserRefreshTokens     []UserRefreshToken
	// UserPushNotification  []UserPushNotification
	// UserFriend            []UserFriend
	// SentFriendRequest     []SentFriendRequest
	// ReceivedFriendRequest []ReceivedFriendRequest
	// BlockedUser           []BlockedUser
	// UserHistory           []UserHistory
}

type UserRefreshToken struct {
	UserID       int64
	RefreshToken uuid.UUID
	IssuedAt     time.Time
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type UserPushNotification struct {
	RegistrationToken string
	Service           NotificationService
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UserHistory struct {
	Action    UserActionType
	CreatedAt time.Time
}
