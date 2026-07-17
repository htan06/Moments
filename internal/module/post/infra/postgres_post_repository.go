package infra

import (
	"context"

	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPostRepository struct {
	conn *pgxpool.Pool
}

func NewPostgresPostRepository(conn *pgxpool.Pool) *PostgresPostRepository {
	return &PostgresPostRepository{
		conn: conn,
	}
}

func (pr *PostgresPostRepository) Create(ctx context.Context, post domain.Post) (int64, error) {
	return 1, nil
}
