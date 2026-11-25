package postgresadapter

import (
	"context"
	"fmt"
)

func (pr *PostgresPostRepository) DeletePostByUserIDAndPostID(ctx context.Context, userID int64, postID int64) error {
	query := `DELETE FROM content.posts WHERE id = $1 AND author_id = $2;`

	_, err := pr.conn.Exec(ctx, query, postID, userID)
	if err != nil {
		return fmt.Errorf("PostgresPostRepository.DeletePostByUserIDAndPostID: %w", err)
	}
	return nil
}

func (p *PostgresPostRepository) DeleteLikePost(ctx context.Context, userID int64, postID int64) error {
	qry := `DELETE FROM content.post_likes WHERE user_id = $1 AND post_id = $2;`
	tag, err := p.conn.Exec(ctx, qry, userID, postID)
	if err != nil {
		return fmt.Errorf("PostgresPostRepository.DeleteLikePost: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("PostgresPostRepository.DeleteLikePost: %s", "Error delete like post")
	}
	return nil
}
