package domain

import "time"

type Follow struct {
	ID          int64
	FollowerID  int64
	FollowingID int64
	CreatedAt   time.Time
}

//read model
type UserSummary struct {
	FollowID  int64     `db:"follow_id" json:"follow_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Name      string    `db:"name" json:"name"`
	Username  string    `db:"username" json:"username"`
	AvatarURL *string    `db:"avatar_id" json:"avatar_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
