package domain

import "time"

type NodeType string
type MediaType string
type Visibility string
type AspectRatio string

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

const (
	Ratio1_1  AspectRatio = "1:1"
	Ratio3_4  AspectRatio = "3:4"
	Ratio3_5  AspectRatio = "3:5"
	Ratio4_3  AspectRatio = "4:3"
	Ratio5_3  AspectRatio = "5:3"
	Ratio16_9 AspectRatio = "16:9"
)

type Content struct {
	Type  NodeType `json:"type"`
	Value string   `json:"value"`
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
	Contents     []Content
	AspectRatio  AspectRatio
	Medias       []Media
	Mentions     []int64
	LikeCount    int
	MediaCount   int
	CommentCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PostPending struct {
	AuthorID    int64       `json:"author_id"`
	Visibility  Visibility  `json:"visibility"`
	Contents    []Content   `json:"contents"`
	AspectRatio AspectRatio `json:"aspect_ratio"`
	MediaIDs    []string    `json:"media_ids"`
	MediaCount  int         `json:"media_count"`
	Mentions    []int64     `json:"mentions"`
}

type UploadPostSession struct {
	MediaIDs []string `json:"media_ids"`
}
