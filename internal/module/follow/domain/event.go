package domain

type PostCreated struct {
	AuthorID int64
	PostID   int64
}

type FollowEventType string

const (
	FollowCreated FollowEventType = "CREATED"
	FollowDeleted FollowEventType = "DELETED"
)

type FollowEvent struct {
	FollowerID  int64           `json:"follower_id"`
	FollowingID int64           `json:"following_id"`
	Type        FollowEventType `json:"type"`
}
