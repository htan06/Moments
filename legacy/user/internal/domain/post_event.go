package domain

import "time"

type PostCreated struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	PostID    int64     `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

type PostDeleted struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	PostID    int64     `json:"post_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
