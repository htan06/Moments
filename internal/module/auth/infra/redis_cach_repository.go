package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/auth/domain"
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

	if _, err := rcr.redisConn.Set(ctx, key, value, ttl).Result(); err != nil {
		return fmt.Errorf("RedisCacheRepository.SetUserPendingIfNotExists: %w", err)
	}

	return nil
}

func (rcr *RedisCacheRepository) GetActiveToken(ctx context.Context, key string) (string, error) {
	token, err := rcr.redisConn.Get(ctx, key).Result()

	if err == redis.Nil {
		return "", errs.NewError(errs.NotFound, nil, domain.UserNotFound)
	}

	return token, nil
}

func (rcr *RedisCacheRepository) SetActiveToken(ctx context.Context, key string, token string, ttl time.Duration) error {
	if _, err := rcr.redisConn.Set(ctx, key, token, ttl).Result(); err != nil {
		return fmt.Errorf("RedisCacheRepository.SetUserPendingIfNotExists: %w", err)
	}

	return nil
}

func (rcr *RedisCacheRepository) Remove(ctx context.Context, key string) error {
	if err := rcr.redisConn.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("RedisCacheRepository.RemoveUserPending: %w", err)
	}
	return nil
}