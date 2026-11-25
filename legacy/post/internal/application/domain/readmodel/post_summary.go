package readmodel

import (
	"time"

	"github.com/htan06/Moments/post/internal/application/domain/vo"
)

type PostSummary struct {
	ID           int64        `json:"id" db:"p_id"`
	ThumbnailID  string       `json:"thumbnail_url" db:"thumbnail_id"`
	Content      []vo.Content `json:"content" db:"content"`
	MediaCount   int          `json:"media_count" db:"media_count"`
	Medias       []Media      `json:"medias" db:"-"`
	LikeCount    int          `json:"like_count" db:"like_count"`
	IsLiked      bool         `json:"is_liked" db:"is_like"`
	CommentCount int          `json:"comment_count" db:"comment_count"`
	CreatedAt    time.Time    `json:"created_at" db:"p_created_at"`
}
