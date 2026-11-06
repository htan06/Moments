package domain

import (
	"encoding/json"
	"time"
)

type FollowCreated struct {
	followerID  int64
	followingID int64
	createdAt   time.Time
}

func NewFollowCreated(followerID, followingID int64, createdAt time.Time) (*FollowCreated, error) {
	if followerID <= 0 || followingID <= 0 {
		return nil, ErrUserIDInvalid
	}
	return &FollowCreated{
		followerID:  followerID,
		followingID: followingID,
		createdAt:   createdAt,
	}, nil
}

func (f *FollowCreated) FollowerID() int64 {
	return f.followerID
}

func (f *FollowCreated) FollowingID() int64 {
	return f.followingID
}

func (f *FollowCreated) CreatedAt() time.Time {
	return f.createdAt
}

func (f *FollowCreated) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			FollowerID  int64     `json:"follower_id"`
			FollowingID int64     `json:"following_id"`
			CreatedAt   time.Time `json:"created_at"`
		}{
			FollowerID:  f.followerID,
			FollowingID: f.followingID,
			CreatedAt:   f.createdAt,
		},
	)
}

type FollowDeleted struct {
	followerID  int64
	followingID int64
	deletedAt   time.Time
}

func NewFollowDeleted(followerID, followingID int64, deletedAt time.Time) (*FollowDeleted, error) {
	if followerID <= 0 || followingID <= 0 {
		return nil, ErrUserIDInvalid
	}
	return &FollowDeleted{
		followerID:  followerID,
		followingID: followingID,
		deletedAt:   deletedAt,
	}, nil
}

func (f *FollowDeleted) FollowerID() int64 {
	return f.followerID
}

func (f *FollowDeleted) FollowingID() int64 {
	return f.followingID
}

func (f *FollowDeleted) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			FollowerID  int64     `json:"follower_id"`
			FollowingID int64     `json:"following_id"`
			DeletedAt   time.Time `json:"deleted_at"`
		}{
			FollowerID:  f.followerID,
			FollowingID: f.followingID,
			DeletedAt:   f.deletedAt,
		},
	)
}
