package postgresadapter

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
)

func (p *PostgresPostRepository) UpadateBatchLikeCount(ctx context.Context, list map[int64]int64) error {
	psqlStatement := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	expr := squirrel.Case()
	for postID, likeCount := range list {
		expr = expr.When(squirrel.Eq{"id": postID}, squirrel.Expr("like_count + ?", likeCount))
	}
	expr = expr.Else("like_count")

	qryBuilder := psqlStatement.Update("content.posts").Set("like_count", expr)
	qry, args, err := qryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("PostgresPostRepository.UpadateBatchLikeCount: %w", err)
	}

	if _, err := p.conn.Exec(ctx, qry, args...); err != nil {
		return fmt.Errorf("PostgresPostRepository.UpadateBatchLikeCount: %w", err)
	}

	// if tag.RowsAffected() != int64(len(list)) {
	// 	return fmt.Errorf("PostgresPostRepository.UpadateBatchLikeCount: %s", "Error update like count")
	// }
	return nil
}
