package connection

import (
	"context"
	"log"
	"user-service/config"

	"github.com/redis/go-redis/v9"
)

func GetRedisConn(cfg *config.RedisConfig) *redis.Client {
	conn := redis.NewClient(
		&redis.Options{
			Addr:     cfg.Address,
			Password: cfg.Password,
		},
	)

	if err := conn.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis connection error: " + err.Error())
	}
	return conn
}
