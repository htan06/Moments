package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func GetRedisConn() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	addr := host + ":" + port

	conn := redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: password,
		},
	)

	if err := conn.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis connection error: " + err.Error())
	}
	return conn
}
