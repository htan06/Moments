package infra

import (
	"context"
	"fmt"
	"strconv"

	"github.com/htan06/Moments/post/internal/errs"
	"github.com/redis/go-redis/v9"
)

func postLikesKey(postID int64) string {
	return fmt.Sprintf("post-likes:%d", postID)
}

type RedisCounterRepository struct {
	conn *redis.Client
	// postRepo domain.PostRepository
}

func NewRedisCounterRepository(
	conn *redis.Client,
	// postRepo domain.PostRepository,
) *RedisCounterRepository {
	return &RedisCounterRepository{
		conn: conn,
		// postRepo: postRepo,
	}
}

func (rc *RedisCounterRepository) IncPostLikes(ctx context.Context, postID int64) error {
	key := postLikesKey(postID)
	if _, err := rc.conn.Exists(ctx, key).Result(); err != nil {
		return errs.NewError(errs.NotFound, err, errs.PostLikeCacheNotFound)
	}

	if _, err := rc.conn.Incr(ctx, key).Result(); err != nil {
		return fmt.Errorf("RedisCounterRepository.IncPostLikes: %w", err)
	}
	return nil
}

func (rc *RedisCounterRepository) SetPostLikesIfNotExists(ctx context.Context, postID int64, likeCount int) error {
	key := postLikesKey(postID)

	if _, err := rc.conn.SetNX(ctx, key, likeCount, 0).Result(); err != nil {
		return fmt.Errorf("RedisCounterRepository.IncPostLikes: %w", err)
	}
	return nil
}

func (rc *RedisCounterRepository) DecPostLikes(ctx context.Context, postID int64) error {
	key := postLikesKey(postID)
	if _, err := rc.conn.Exists(ctx, key).Result(); err != nil {
		return errs.NewError(errs.NotFound, err, errs.PostLikeCacheNotFound)
	}

	if _, err := rc.conn.Decr(ctx, key).Result(); err != nil {
		return fmt.Errorf("RedisCounterRepository.IncPostLikes: %w", err)
	}
	return nil
}

func (rc *RedisCounterRepository) GetLikeCount(ctx context.Context, postIDs ...int64) (map[int64]int64, error) {
	len := len(postIDs)
	if len == 0 {
		return map[int64]int64{}, nil
	}
	var keys []string

	for _, id := range postIDs {
		keys = append(keys, postLikesKey(id))
	}

	vals, err := rc.conn.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("RedisCounterRepository.IncPostLikes: %w", err)
	}

	postLikes := make(map[int64]int64, 0)
	i := 0
	for i < len {
		if vals[i] != nil {
			likeCount, err := strconv.ParseInt(vals[i].(string), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("RedisCounterRepository.IncPostLikes: %w", err)
			}
			postLikes[postIDs[i]] = likeCount
		}
		i++
	}
	return postLikes, nil
}
