package model

type UserProfile struct {
	Username      string
	FirstName     string
	LastName      *string
	AvatarURL     *string
	CoverPhotoURL *string
}