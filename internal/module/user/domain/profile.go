package domain

type UserProfile struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	AvatarURL      *string `json:"avatar_url"`
	Bio            *string `json:"bio"`
	FollowersCount int64  `json:"followers_count"`
	FollowingCount int64  `json:"following_count"`
	PostsCount     int64  `json:"posts_count"`
}
