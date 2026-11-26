// Package redisx provides the Redis client.
package redisx

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string
	Password string
}

// Connect creates a client and verifies it with a ping.
func Connect(ctx context.Context, cfg Config) (*redis.Client, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})
	if err := c.Ping(ctx).Err(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("redisx: ping: %w", err)
	}
	return c, nil
}
