package entity

import (
	"regexp"
	"time"
	"user-service/internal/application/domain"
	"user-service/internal/errs"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

var (
	validateUsername = func(username string) error {
		if !usernameRegex.MatchString(username) {
			return errs.NewError(errs.Invalid, nil, errs.UsernameInvalid)
		}
		return nil
	}

	validateName = func(name string) error {
		if len(name) < 2 || len(name) > 100 {
			return errs.NewError(errs.Invalid, nil, errs.NameInvalid)
		}
		return nil
	}
)

type Profile struct {
	userID             int64     `db:"user_id"`
	username           string    `db:"username"`
	name               string    `db:"name"`
	avatarKey          *string   `db:"avatar_key"`
	avatarThumbnailKey *string   `db:"avatar_thumbnail_key"`
	bio                *string   `db:"bio"`
	followersCount     int32     `db:"followers_count"`
	followingCount     int32     `db:"following_count"`
	postsCount         int32     `db:"posts_count"`
	createdAt          time.Time `db:"created_at"`
	updatedAt          time.Time `db:"updated_at"`
}

func NewProfile(userID int64, name string, username string) (*Profile, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validateUsername(username); err != nil {
		return nil, err
	}

	return &Profile{
		userID:         userID,
		name:           name,
		username:       username,
		followersCount: 0,
		followingCount: 0,
		postsCount:     0,
		createdAt:      time.Now().UTC(),
		updatedAt:      time.Now().UTC(),
	}, nil
}

func (p *Profile) UserID() int64 {
	return p.userID
}

func (p *Profile) Username() string {
	return p.username
}

func (p *Profile) ChangeUsername(username string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	p.username = username
	return nil
}

func (p *Profile) Name() string {
	return p.name
}

func (p *Profile) ChangeName(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	p.name = name
	return nil
}

func (p *Profile) AvatarKey() *string {
	return p.avatarKey
}

func (p *Profile) ChangeAvatarKey(avatarKey *string) {
	p.avatarKey = avatarKey
}

func (p *Profile) AvatarThumbnailKey() *string {
	return p.avatarThumbnailKey
}

func (p *Profile) ChangeAvatarThumbnailKey(avatarThumbnailKey *string) {
	p.avatarThumbnailKey = avatarThumbnailKey
}

func (p *Profile) Bio() *string {
	return p.bio
}

func (p *Profile) UpdateBio(bio *string) {
	p.bio = bio
}

func (p *Profile) FollowersCount() int32 {
	return p.followersCount
}

func (p *Profile) IncFollowersCount(n int32) error {
	if n <= 0 {
		return errs.NewError(errs.Invalid, nil, domain.ErrNegativeCount)
	}
	p.followersCount += n
	return nil
}

func (p *Profile) DecFollowersCount(n int32) error {
	if n <= 0 || p.followersCount-n < 0 {
		return errs.NewError(errs.Invalid, nil, domain.ErrNegativeCount)
	}
	p.followersCount -= n
	return nil
}

func (p *Profile) FollowingCount() int32 {
	return p.followingCount
}

func (p *Profile) IncFollowingCount(n int32) error {
	if n <= 0 {
		return errs.NewError(errs.Invalid, nil, domain.ErrNegativeCount)
	}
	p.followingCount += n
	return nil
}

func (p *Profile) DecFollowingCount(n int32) error {
	if n <= 0 || p.followingCount-n < 0 {
		return errs.NewError(errs.Invalid, nil, domain.ErrNegativeCount)
	}
	p.followingCount -= n
	return nil
}

func (p *Profile) PostsCount() int32 {
	return p.postsCount
}

func (p *Profile) IncPostsCount(n int32) error {
	if n <= 0 {
		return errs.NewError(errs.Invalid, nil, domain.ErrNegativeCount)
	}
	p.postsCount += n
	return nil
}

func (p *Profile) DecPostsCount(n int32) error {
	if n <= 0 || p.postsCount-n < 0 {
		return errs.NewError(errs.Invalid, nil, domain.ErrNegativeCount)
	}
	p.postsCount -= n
	return nil
}

func (p *Profile) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Profile) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *Profile) SetUpdatedAt(updatedAt time.Time) {
	p.updatedAt = updatedAt
}
