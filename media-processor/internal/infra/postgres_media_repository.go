package infra

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMediarepository struct {
	postgresConn *pgxpool.Pool
}

func NewPostgresMediarepository(postgresConn *pgxpool.Pool) *PostgresMediarepository {
	return &PostgresMediarepository{
		postgresConn: postgresConn,
	}
}

func (p *PostgresMediarepository) Update(ctx context.Context, id int64, mediaID string) error {
	qry := `UPDATE content.medias SET media_id = $1 WHERE id = $2`
	cmd, err := p.postgresConn.Exec(ctx, qry, mediaID, id)
	if err != nil {
		return fmt.Errorf("PostgresMediarepository.Update: %w", err)
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("PostgresMediarepository.Update: Error update")
	}
	return nil
}
