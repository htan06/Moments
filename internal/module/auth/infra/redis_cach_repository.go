package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/htan06/echo-messenger-rest-api/internal/errs"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/domain"
	"github.com/redis/go-redis/v9"
)

type RedisCacheRepository struct {
	redisConn *redis.Client
}

func NewRedisCacheRepository(redisConn *redis.Client) *RedisCacheRepository {
	return &RedisCacheRepository{
		redisConn: redisConn,
	}
}

func (rcr *RedisCacheRepository) GetUserPending(ctx context.Context, key string) (domain.UserPending, error) {
	data, err := rcr.redisConn.Get(ctx, key).Result()

	if err == redis.Nil {
		return domain.UserPending{}, errs.NewError(errs.NotFound, nil, domain.UserNotFound)
	}

	var val domain.UserPending
	if err := json.Unmarshal([]byte(data), &val); err != nil {
		return domain.UserPending{}, fmt.Errorf("RedisCacheRepository.GetUserPending: %w", err)
	}
	return val, nil
}

func (rcr *RedisCacheRepository) SetUserPendingIfNotExists(ctx context.Context, key string, u domain.UserPending, ttl time.Duration) error {
	value, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("RedisCacheRepository.SetUserPendingIfNotExists: %w", err)
	}

	if _, err := rcr.redisConn.SetNX(ctx, key, value, ttl).Result(); err != nil {
		return fmt.Errorf("RedisCacheRepository.SetUserPendingIfNotExists: %w", err)
	}

	return nil
}

func (rcr *RedisCacheRepository) RemoveUserPending(ctx context.Context, key string) error {
	if err := rcr.redisConn.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("RedisCacheRepository.RemoveUserPending: %w", err)
	}
	return nil
}
