package model

import "time"

type FriendRequest struct {
	ID         int64     `json:"id"`
	SenderID   int64     `json:"sender_id"`
	ReceiverID int64     `json:"receiver_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type FriendRequestProfile struct {
	ID int64 `json:"id"`
	UserProfile
}