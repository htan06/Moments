package domain

import "time"

type NodeType string
type MediaType string
type Visibility string

const (
	Text    NodeType = "TEXT"
	Mention NodeType = "MENTION"
	Hashtag NodeType = "HASHTAG"
)

const (
	Image MediaType = "IMAGE"
	Video MediaType = "VIDEO"
)

const (
	Private Visibility = "PRIVATE"
	Public  Visibility = "PUBLIC"
	Friend  Visibility = "FRIEND"
)

type Node struct {
	Type  NodeType
	Value any
}

type Media struct {
	ID           int64
	MediaID      string
	Type         MediaType
	DisplayOrder int
	Width        int
	Height       int
	Duration     time.Duration
	Size         int
}

type Post struct {
	ID           int64
	AuthorID     int64
	Visibility   Visibility
	Content      []Node
	Medias       []Media
	LikeCount    int
	MediaCount   int
	CommentCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
