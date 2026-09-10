package domain

type PostEventType string
type InteractionType string
type Action string

const (
	Created Action = "CREATED"
	Deleted Action = "DELETED"

	LikeInteraction    InteractionType = "LIKE"
	CommentInteraction InteractionType = "COMMENT"
	RepostInteraction  InteractionType = "REPOST"
)

type PostEvent struct {
	AuthorID int64  `json:"author_id"`
	PostID   int64  `json:"post_id"`
	Action   Action `json:"action"`
}

type InteractionEvent struct {
	UserID          int64           `json:"user_id"`
	PostID          int64           `json:"post_id"`
	TypeInteraction InteractionType `json:"type"`
	Action          Action          `json:"action"`
}
