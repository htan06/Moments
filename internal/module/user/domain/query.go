package domain

type ProfileQry struct {
	ID             string  `json:"id" db:"id"`
	Username       string  `json:"username" db:"username"`
	Name           string  `json:"name" db:"name"`
	AvatarURL      *string `json:"avatar_url" db:"avatar_id"`
	Bio            *string `json:"bio" db:"bio"`
	FollowersCount int64   `json:"followers_count" db:"followers_count"`
	FollowingCount int64   `json:"following_count" db:"following_count"`
	PostsCount     int64   `json:"posts_count" db:"posts_count"`
}
