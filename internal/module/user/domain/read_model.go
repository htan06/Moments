package domain

type ProfileReadModel struct {
	IsOwner            bool          `json:"is_owner" db:"-"`
	UserID             int64         `json:"user_id" db:"user_id"`
	Username           string        `json:"username" db:"username"`
	Name               string        `json:"name" db:"name"`
	AvatarURL          *string       `json:"avatar_url" db:"avatar_id"`
	AvatarThumbnailURL *string       `json:"avatar_thumbnail_url" db:"avatar_thumbnail_id"`
	Bio                *string       `json:"bio" db:"bio"`
	FollowersCount     int64         `json:"followers_count" db:"followers_count"`
	FollowingCount     int64         `json:"following_count" db:"following_count"`
	PostsCount         int64         `json:"posts_count" db:"posts_count"`
	Relationship       *Relationship `json:"relationship,omitempty" db:"relationship"`
}

type Relationship struct {
	FollowId *int64           `json:"follow_id" db:"follow_id"`
	Type     RelationshipType `json:"type" db:"type"`
}

type RelationshipType string

const (
	Blocked   RelationshipType = "BLOCKED"
	Following RelationshipType = "FOLLOWING"
	None      RelationshipType = "NONE"
)

type ProfileSummaryReadModel struct {
	UserID             int64   `json:"user_id" db:"user_id"`
	Username           string  `json:"username" db:"username"`
	Name               string  `json:"name" db:"name"`
	AvatarThumbnailURL *string `json:"avatar_thumbnail_url" db:"avatar_thumbnail_id"`
}
