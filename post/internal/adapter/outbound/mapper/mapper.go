package mapper

import (
	"fmt"

	"github.com/htan06/Moments/post/internal/application/domain/readmodel"
	"github.com/jackc/pgx/v5"
)

func RowsToPosts(rows pgx.Rows) ([]readmodel.Post, error) {

	postIndex := map[int64]int{}
	posts := make([]readmodel.Post, 0)
	for rows.Next() {
		post := readmodel.Post{}
		media := readmodel.Media{}

		if err := rows.Scan(
			&post.ID, &post.AuthorName, &post.AuthorUsername, &post.AuthorAvatarThumbnailID, &post.ThumbnailID, &post.Content, &post.MediaCount, &post.LikeCount, &post.CommentCount, &post.CreatedAt,
			&media.ID, &media.Type, &media.MediaID, &media.DisplayOrder, &media.Width, &media.Height, &media.Duration, &media.Size, &media.CreatedAt, &post.IsLiked,
		); err != nil {
			return []readmodel.Post{}, fmt.Errorf("RowsToPosts: %w", err)
		}

		index, exists := postIndex[post.ID]
		if !exists {
			posts = append(posts, post)
			index = len(posts) - 1
			postIndex[post.ID] = index
		}

		posts[index].Medias = append(posts[index].Medias, media)
	}

	return posts, nil
}
