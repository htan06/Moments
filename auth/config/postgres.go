package config

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetPostgresConn() *pgxpool.Pool {
	databaseUrl := os.Getenv("POSTGRES_DATABASE_URL")
	
	conn, err := pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatal("Postgres connection error: " + err.Error())
	}

	if err := conn.Ping(context.Background()); err != nil {
		log.Fatal("Postgres connection error: " + err.Error())
	}
	return conn
}