package post

import "github.com/htan06/Moments/post/internal/domain"

// Created is emitted once a new post has its id (AssignID).
type Created struct {
	domain.EventMeta
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}

func (Created) EventName() string { return "post.created" }

// Published is emitted when the last media is processed and the post
// becomes visible.
type Published struct {
	domain.EventMeta
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}

func (Published) EventName() string { return "post.published" }

type ContentUpdated struct {
	domain.EventMeta
	PostID   int64 `json:"post_id"`
	AuthorID int64 `json:"author_id"`
}

func (ContentUpdated) EventName() string { return "post.content_updated" }

type VisibilityUpdated struct {
	domain.EventMeta
	PostID     int64      `json:"post_id"`
	AuthorID   int64      `json:"author_id"`
	Visibility Visibility `json:"visibility"`
}

func (VisibilityUpdated) EventName() string { return "post.visibility_updated" }

// Deleted carries the media keys so their objects can be removed.
type Deleted struct {
	domain.EventMeta
	PostID    int64    `json:"post_id"`
	AuthorID  int64    `json:"author_id"`
	MediaKeys []string `json:"media_keys"`
}

func (Deleted) EventName() string { return "post.deleted" }

var (
	_ domain.Event = Created{}
	_ domain.Event = Published{}
	_ domain.Event = ContentUpdated{}
	_ domain.Event = VisibilityUpdated{}
	_ domain.Event = Deleted{}
)
