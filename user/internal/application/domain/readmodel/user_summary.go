package readmodel

import "time"

type ProfileSummary struct {
	UserID             int64   `json:"user_id" db:"user_id"`
	Username           string  `json:"username" db:"username"`
	Name               string  `json:"name" db:"name"`
	AvatarThumbnailURL *string `json:"avatar_thumbnail_url" db:"avatar_thumbnail_key"`
}

type UserSummary struct {
	FollowID           int64     `db:"follow_id" json:"follow_id"`
	UserID             int64     `db:"user_id" json:"user_id"`
	Name               string    `db:"name" json:"name"`
	Username           string    `db:"username" json:"username"`
	AvatarThumbnailURL *string   `db:"avatar_thumbnail_key" json:"avatar_thumbnail_url"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}
