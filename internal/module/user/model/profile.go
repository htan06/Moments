package model

type UserProfile struct {
	ID            int64   `json:"id"`
	Username      string  `json:"username"`
	FirstName     string  `json:"first_name"`
	LastName      *string `json:"last_name,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	CoverPhotoURL *string `json:"cover_photo_url,omitempty"`
}