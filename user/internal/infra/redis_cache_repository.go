package infra

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCacheRepository struct {
	conn *redis.Client
}

func NewRedisCacheRepository(conn *redis.Client) *RedisCacheRepository {
	return &RedisCacheRepository{
		conn: conn,
	}
}

func (rcr *RedisCacheRepository) SetUploadAvatarSession(ctx context.Context, key string, avatarID string, ttl time.Duration) error {
	return rcr.conn.SetNX(ctx, key, avatarID, ttl).Err()
}

func (rcr *RedisCacheRepository) GetUploadAvatarSession(ctx context.Context, key string) (string, error) {
	avatarID, err := rcr.conn.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return avatarID, nil
}

func (rcr *RedisCacheRepository) RemoveSession(ctx context.Context, key string) error {
	if err := rcr.conn.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}
