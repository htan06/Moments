package entity

import (
	"time"
	"user-service/internal/application/domain"
	"user-service/internal/errs"
)

type Follow struct {
	followerID  int64
	followingID int64
	createdAt   time.Time
}

func NewFollow(followerID, followingID int64) (*Follow, error) {
	if followerID <= 0 || followingID <= 0 {
		return nil, errs.NewError(errs.Invalid, nil, domain.ErrUserIDInvalid)
	}
	return &Follow{
		followerID:  followerID,
		followingID: followingID,
	}, nil
}

func (f *Follow) FollowerID() int64 {
	return f.followerID
}

func (f *Follow) FollowingID() int64 {
	return f.followingID
}

func (f *Follow) CreatedAt() time.Time {
	return f.createdAt
}
