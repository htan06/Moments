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
