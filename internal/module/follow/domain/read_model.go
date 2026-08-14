package domain

import "time"

type UserSummary struct {
	FollowID           int64     `db:"follow_id" json:"follow_id"`
	UserID             int64     `db:"user_id" json:"user_id"`
	Name               string    `db:"name" json:"name"`
	Username           string    `db:"username" json:"username"`
	AvatarThumbnailURL *string   `db:"avatar_thumbnail_id" json:"avatar_thumbnail_url"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}
