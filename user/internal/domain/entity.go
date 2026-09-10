package domain

import (
	"regexp"
	"time"
	"user-service/internal/errs"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
)

type Profile struct {
	UserID            int64     `db:"user_id"`
	Username          string    `db:"username"`
	Name              string    `db:"name"`
	AvatarID          *string   `db:"avatar_id"`
	AvatarThumbnailID *string   `db:"avatar_thumbnail_id"`
	Bio               *string   `db:"bio"`
	FollowersCount    int64     `db:"followers_count"`
	FollowingCount    int64     `db:"following_count"`
	PostsCount        int64     `db:"posts_count"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

func NewProfile(userID int64, name string, username string) (*Profile, error) {
	if len(name) < 5 {
		return nil, errs.NewError(errs.Invalid, nil, errs.NameInvalid)
	}

	if !usernameRegex.MatchString(username) {
		return nil, errs.NewError(errs.Invalid, nil, errs.UsernameInvalid)
	}

	return &Profile{
		UserID:         userID,
		Name:           name,
		Username:       username,
		FollowersCount: 0,
		FollowingCount: 0,
		PostsCount:     0,
	}, nil
}

type Follow struct {
	ID          int64
	FollowerID  int64
	FollowingID int64
	CreatedAt   time.Time
}
