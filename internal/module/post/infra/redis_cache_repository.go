package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
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

func (rc *RedisCacheRepository) SetPostPending(ctx context.Context, postSessionID string, postPending domain.PostPending) error {
	data, err := json.Marshal(postPending)
	if err != nil {
		return fmt.Errorf("RedisCacheRepository.SetPostPending: %w", err)
	}

	key := fmt.Sprintf("upload-post-session:%s", postSessionID)
	cmd := rc.conn.Set(ctx, key, data, time.Minute*5)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("RedisCacheRepository.SetPostPending: %w", err)
	}
	return nil
}

func (rc *RedisCacheRepository) GetPostPending(ctx context.Context, postSessionID string) (domain.PostPending, error) {
	key := fmt.Sprintf("upload-post-session:%s", postSessionID)
	data, err := rc.conn.Get(ctx, key).Bytes()
	if err != nil {
		return domain.PostPending{}, fmt.Errorf("RedisCacheRepository.SetPostPending: %w", err)
	}

	var postPending domain.PostPending
	if err := json.Unmarshal(data, &postPending); err != nil {
		return domain.PostPending{}, fmt.Errorf("RedisCacheRepository.SetPostPending: %w", err)
	}

	return postPending, nil
}
