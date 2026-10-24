package event

type PostEvent struct {
	AuthorID int64  `json:"author_id"`
	PostID   int64  `json:"post_id"`
	Action   Action `json:"action"`
}