package infra

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	conn *pgxpool.Pool
}

func NewPostgresUserRepository(conn *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		conn: conn,
	}
}

func (ur *PostgresUserRepository) GetIDByUsername(ctx context.Context, username string) (*int64, error) {
	query := `SELECT user_id from profile.users WHERE username = $1;`

	var id int64
	if err := ur.conn.QueryRow(ctx, query, username).Scan(&id); err != nil {
		return nil, fmt.Errorf("PostgresUserRepository.GetIDByUsername: %w", err)
	}
	return &id, nil
}
