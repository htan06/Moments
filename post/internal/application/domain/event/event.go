package event

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

type InteractionEvent struct {
	UserID          int64           `json:"user_id"`
	PostID          int64           `json:"post_id"`
	TypeInteraction InteractionType `json:"type"`
	Action          Action          `json:"action"`
}

type ProcessMediaJob struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Bucket      string `json:"bucket"`
	MediaType   string `json:"media_type"`
	AspectRatio string `json:"aspect_ratio"`
}
