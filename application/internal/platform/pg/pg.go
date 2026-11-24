// Package pg provides the PostgreSQL connection pool and error helpers.
package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	URI      string
	MaxConns int32
	MinConns int32
}

// Connect opens a pool and verifies it with a ping.
func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(cfg.URI)
	if err != nil {
		// Do not wrap: the parse error may echo the URI (with password).
		return nil, errors.New("pg: invalid connection uri")
	}
	pc.MaxConns = cfg.MaxConns
	pc.MinConns = cfg.MinConns

	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("pg: create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg: ping: %w", err)
	}
	return pool, nil
}

const uniqueViolation = "23505"

// IsUniqueViolation reports whether err is a unique violation on the given
// constraint (index) name.
func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == uniqueViolation &&
		pgErr.ConstraintName == constraint
}
