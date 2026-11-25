package readmodel

import (
	"time"

	"github.com/htan06/Moments/post/internal/application/domain/vo"
)

type PostAuth struct {
	ID                      int64        `json:"id" db:"p_id"`
	AuthorID                int64        `json:"author_id" db:"author_id"`
	AuthorName              string       `json:"author_name" db:"author_name"`
	AuthorUsername          string       `json:"author_username" db:"author_username"`
	AuthorAvatarThumbnailID *string      `json:"author_avatar_thumbnail_url" db:"author_avatar_thumbnail_id"`
	ThumbnailID             string       `json:"thumbnail_url" db:"thumbnail_id"`
	Content                 []vo.Content `json:"content" db:"content"`
	MediaCount              int          `json:"media_count" db:"media_count"`
	Medias                  []Media      `json:"medias" db:"-"`
	LikeCount               int          `json:"like_count" db:"like_count"`
	IsLiked                 bool         `json:"is_liked" db:"is_like"`
	CommentCount            int          `json:"comment_count" db:"comment_count"`
	CreatedAt               time.Time    `json:"created_at" db:"p_created_at"`
}
