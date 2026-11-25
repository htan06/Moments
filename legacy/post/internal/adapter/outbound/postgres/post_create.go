package postgresadapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/htan06/Moments/post/internal/adapter/outbound/postgres/sqlc/queries"
	"github.com/htan06/Moments/post/internal/domain/entity"
)

func (pr *PostgresPostRepository) CreatePost(ctx context.Context, post entity.Post) (*int64, error) {
	tx, err := pr.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}
	defer tx.Rollback(ctx)
	queriesTx := pr.queries.WithTx(tx)

	contentBytes, err := json.Marshal(post.Content())
	if err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}

	postID, err := queriesTx.CreatePost(ctx, queries.CreatePostParams{
		AuthorID:     post.AuthorID(),
		Content:      contentBytes,
		Visibility:   queries.ContentVisibility(post.Visibility()),
		ThumbnailKey: post.ThumbnailKey(),
		MediaCount:   post.MediaCount(),
		AspectRatio:  queries.ContentAspectRatio(post.AspectRatio()),
		Status:       queries.ContentPostStatus(post.Status()),
	})
	if err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}

	mediasParams := []queries.CreateMediasParams{}
	for _, m := range post.Medias() {
		mediasParams = append(mediasParams, queries.CreateMediasParams{
			PostID:       postID,
			Type:         queries.ContentMediaType(m.Type()),
			MediaKey:     m.MediaKey(),
			DisplayOrder: m.DisplayOrder(),
			Width:        m.Width(),
			Height:       m.Height(),
		})
	}
	if err := queriesTx.CreateMedias(ctx, mediasParams).Close(); err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}

	mentionsParams := []queries.CreateMentionsParams{}
	for _, mID := range post.MentionIDs() {
		mentionsParams = append(mentionsParams, queries.CreateMentionsParams{
			PostID: postID,
			UserID: mID,
		})
	}
	if err := queriesTx.CreateMentions(ctx, mentionsParams).Close(); err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("PostgresPostRepository.Create: %w", err)
	}
	return &postID, nil
}

func (p *PostgresPostRepository) CreateLikePost(ctx context.Context, userID int64, postID int64) error {
	qry := `INSERT INTO content.post_likes (user_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	tag, err := p.conn.Exec(ctx, qry, userID, postID)

	if err != nil {
		return fmt.Errorf("PostgresPostRepository.CreateLikePost: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("PostgresPostRepository.CreateLikePost: %s", "Error insert like post")
	}
	return nil
}
