package readmodel

type Profile struct {
	IsOwner            bool    `json:"is_owner" db:"-"`
	UserID             int64   `json:"user_id" db:"user_id"`
	Username           string  `json:"username" db:"username"`
	Name               string  `json:"name" db:"name"`
	AvatarURL          *string `json:"avatar_url" db:"avatar_key"`
	AvatarThumbnailURL *string `json:"avatar_thumbnail_url" db:"avatar_thumbnail_key"`
	Bio                *string `json:"bio" db:"bio"`
	FollowingCount     int32   `json:"following_count" db:"following_count"`
	FollowersCount     int32   `json:"followers_count" db:"followers_count"`
	PostsCount         int32   `json:"posts_count" db:"posts_count"`
	Relationship       string  `json:"relationship,omitempty" db:"relationship"`
}
