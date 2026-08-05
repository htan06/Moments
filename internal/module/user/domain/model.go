package domain

type Profile struct {
	id             string
	username       string
	name           string
	avatarID       *string
	bio            *string
	followersCount int64
	followingCount int64
	postsCount     int64
}

type ProfileReadModel struct {
	IsOwner        bool          `json:"is_owner" db:"-"`
	UserID         int64         `json:"user_id" db:"user_id"`
	Username       string        `json:"username" db:"username"`
	Name           string        `json:"name" db:"name"`
	AvatarURL      *string       `json:"avatar_url" db:"avatar_id"`
	Bio            *string       `json:"bio" db:"bio"`
	FollowersCount int64         `json:"followers_count" db:"followers_count"`
	FollowingCount int64         `json:"following_count" db:"following_count"`
	PostsCount     int64         `json:"posts_count" db:"posts_count"`
	Relationship   *Relationship `json:"relationship,omitempty" db:"relationship"`
}

type Relationship string

const (
	Blocked   Relationship = "BLOCKED"
	Following Relationship = "FOLLOWING"
	None      Relationship = "NONE"
)
