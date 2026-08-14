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
	conn *redis.Client
}

func NewRedisCacheRepository(conn *redis.Client) *RedisCacheRepository {
	return &RedisCacheRepository{
		conn: conn,
	}
}

func (rc *RedisCacheRepository) SetUploadPostSession(ctx context.Context, key string, uploadPostSession domain.UploadPostSession) error {
	data, err := json.Marshal(uploadPostSession)
	if err != nil {
		return fmt.Errorf("RedisCacheRepository.SetUploadPostSession: %w", err)
	}

	cmd := rc.conn.Set(ctx, key, data, time.Minute*5)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("RedisCacheRepository.SetUploadPostSession: %w", err)
	}
	return nil
}

func (rc *RedisCacheRepository) GetUploadPostSession(ctx context.Context, key string) (domain.UploadPostSession, error) {
	data, err := rc.conn.Get(ctx, key).Bytes()
	if err != nil {
		return domain.UploadPostSession{}, fmt.Errorf("RedisCacheRepository.GetUploadPostSession: %w", err)
	}

	var uploadPostSession domain.UploadPostSession
	if err := json.Unmarshal(data, &uploadPostSession); err != nil {
		return domain.UploadPostSession{}, fmt.Errorf("RedisCacheRepository.GetUploadPostSession: %w", err)
	}

	return uploadPostSession, nil
}
