package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
		return domain.UserPending{}, errors.New("Key not exists")
	}

	var val domain.UserPending
	if err := json.Unmarshal([]byte(data), &val); err != nil {
		return domain.UserPending{}, err
	}
	return val, nil
}

func (rcr *RedisCacheRepository) SetUserPendingIfNotExists(ctx context.Context, key string, u domain.UserPending, ttl time.Duration) error {
	value, err := json.Marshal(u)
	if err != nil {
		return err
	}

	ok, err := rcr.redisConn.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return fmt.Errorf("RedisCacheRepository[Set]: %w", err)
	}

	if ok {
		return nil
	}
	return errors.New("RedisCacheRepository[Set]: " + "Key already exists")
}

func (rcr *RedisCacheRepository) RemoveUserPending(ctx context.Context, key string) error {
	if err := rcr.redisConn.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("RedisCacheRepository[Remove]: %w", err)
	}
	return nil
}