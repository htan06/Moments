package connection

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetPostgresConn(cfg *pgxpool.Config) *pgxpool.Pool {
	conn, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Fatal("Postgres connection error: " + err.Error())
	}

	if err := conn.Ping(context.Background()); err != nil {
		log.Fatal("Postgres connection error: " + err.Error())
	}
	return conn
}
