package readmodel

import (
	"time"

	"github.com/htan06/Moments/post/internal/domain/entity"
)

type Post struct {
	ID                      int64             `json:"id" db:"id"`
	AuthorID                int64             `json:"author_id" db:"author_id"`
	AuthorName              string            `json:"author_name" db:"author_name"`
	AuthorUsername          string            `json:"author_username" db:"author_username"`
	AuthorAvatarThumbnailID *string           `json:"author_avatar_thumbnail_url" db:"author_avatar_thumbnail_id"`
	Content                 Content           `json:"content" db:"content"`
	Visibility              entity.Visibility `json:"visibility" db:"visibility"`
	ThumbnailID             string            `json:"thumbnail_url" db:"thumbnail_key"`
	MediaCount              int               `json:"media_count" db:"media_count"`
	Medias                  []Media           `json:"medias" db:"-"`
	AspectRatio             string            `json:"aspect_ratio" db:"aspect_ratio"`
	LikeCount               int               `json:"like_count" db:"like_count"`
	CommentCount            int               `json:"comment_count" db:"comment_count"`
	IsLiked                 bool              `json:"is_liked" db:"is_like"`
	CreatedAt               time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time         `json:"updated_at" db:"updated_at"`
}

type Media struct {
	ID           int64     `json:"id" db:"m_id"`
	Type         string    `json:"type" db:"type"`
	MediaID      string    `json:"media_url" db:"media_key"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	Width        *int      `json:"width,omitempty" db:"width"`
	Height       *int      `json:"height,omitempty" db:"height"`
	Duration     *int      `json:"duration,omitempty" db:"duration"`
	Size         *int64    `json:"size,omitempty" db:"size"`
	CreatedAt    time.Time `json:"created_at" db:"m_created_at"`
}
