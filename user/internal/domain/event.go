package domain

type Action string

const (
	Created Action = "CREATED"
	Deleted Action = "DELETED"
)

type PostEvent struct {
	AuthorID int64  `json:"author_id"`
	PostID   int64  `json:"post_id"`
	Action   Action `json:"action"`
}

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
