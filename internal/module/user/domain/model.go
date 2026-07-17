package domain

import (
	"regexp"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{6,32}}$`)
)

type Profile struct {
	id             string
	username       string
	name           string
	avatarID       *string
	bio            *string
	followersCount int64
	followingCount int64
	postsCount     int64
}

func NewProfile(
	id string,
	username string,
	name string,
	avatarID *string,
	bio *string,
	followersCount int64,
	followingCount int64,
	postsCount int64,
) (*Profile, error) {
	if !usernameRegex.MatchString(username) {
		return &Profile{}, errs.NewError(errs.Invalid, nil, errs.UsernameInvalid)
	}
	return &Profile{
		id:             id,
		username:       username,
		name:           name,
		avatarID:       avatarID,
		bio:            bio,
		followersCount: followersCount,
		followingCount: followingCount,
		postsCount:     postsCount,
	}, nil
}

func (p *Profile) UpdateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return errs.NewError(errs.Invalid, nil, errs.UsernameInvalid)
	}
	p.username = username
	return nil
}

func (p *Profile) UpdateName(name string) {
	p.name = name
}

func (p *Profile) UpdateAvatar(avatarID *string) {
	p.avatarID = avatarID
}

func (p *Profile) UpdateBio(bio *string) {
	p.bio = bio
}

func (p *Profile) ID() string {
	return p.id
}

func (p *Profile) Username() string {
	return p.username
}

func (p *Profile) Name() string {
	return p.name
}

func (p *Profile) AvatarID() *string {
	return p.avatarID
}

func (p *Profile) Bio() *string {
	return p.bio
}

func (p *Profile) FollowersCount() int64 {
	return p.followersCount
}

func (p *Profile) FollowingCount() int64 {
	return p.followingCount
}

func (p *Profile) PostsCount() int64 {
	return p.postsCount
}
