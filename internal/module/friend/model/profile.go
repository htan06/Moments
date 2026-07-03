package model

type UserProfile struct {
	UserID        int64   `json:"user_id"`
	Username      string  `json:"username"`
	FirstName     string  `json:"first_name"`
	LastName      *string `json:"last_name"`
	AvatarURL     *string `json:"avatar_url"`
	CoverPhotoURL *string `json:"cover_photo_url"`
}
