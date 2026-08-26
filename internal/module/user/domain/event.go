package domain

type PostEventType string

const (
	Created PostEventType = "CREATED"
	Deleted PostEventType = "DELETED"
)

type PostEvent struct {
	AuthorID int64         `json:"author_id"`
	PostID   int64         `json:"post_id"`
	Type     PostEventType `json:"type"`
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
