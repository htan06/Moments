package user

type UpdateInfoReq struct {
	FirstName     string  `json:"first_name"`
	LastName      *string `json:"last_name,omitempty"`
	AvatarURL     string  `json:"avatar_url"`
	CoverPhotoURL string  `json:"cover_photo_url"`
}

type ChangeReadStatusReq struct {
	Status bool `json:"status"`
}

type UpdateUsernameReq struct {
	Username string `json:"username"`
}