package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/htan06/Moments/internal/module/post/domain"
	"github.com/redis/go-redis/v9"
)

type RedisCacheRepository struct {
	conn     *redis.Client
	// postRepo domain.PostRepository
}

func NewRedisCacheRepository(
	conn *redis.Client,
	// postRepo domain.PostRepository,
) *RedisCacheRepository {
	return &RedisCacheRepository{
		conn:     conn,
		// postRepo: postRepo,
	}
}

func (rc *RedisCacheRepository) SetUploadPostSession(ctx context.Context, key string, createPostSession *domain.CreatePostSession) error {
	data, err := json.Marshal(*createPostSession)
	if err != nil {
		return fmt.Errorf("RedisCacheRepository.SetUploadPostSession: %w", err)
	}

	cmd := rc.conn.Set(ctx, key, data, time.Minute*5)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("RedisCacheRepository.SetUploadPostSession: %w", err)
	}
	return nil
}

func (rc *RedisCacheRepository) GetUploadPostSession(ctx context.Context, key string) (domain.CreatePostSession, error) {
	data, err := rc.conn.Get(ctx, key).Bytes()
	if err != nil {
		return domain.CreatePostSession{}, fmt.Errorf("RedisCacheRepository.GetUploadPostSession: %w", err)
	}

	var uploadPostSession domain.CreatePostSession
	if err := json.Unmarshal(data, &uploadPostSession); err != nil {
		return domain.CreatePostSession{}, fmt.Errorf("RedisCacheRepository.GetUploadPostSession: %w", err)
	}

	return uploadPostSession, nil
}

func (rc *RedisCacheRepository) IncPostLikes(ctx context.Context, key string) error {
	// if _, err := rc.conn.Incr(ctx, key).Result(); err != nil && err == redis.Nil {

	// }
	return nil
}
func (rc *RedisCacheRepository) DecPostLikes(ctx context.Context, key string) error {
	return nil
}
