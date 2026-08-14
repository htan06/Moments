package domain

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