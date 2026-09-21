package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/errs"
	"github.com/redis/go-redis/v9"
)

func createPostSessionKey(userID int64, sessionID string) string {
	return fmt.Sprintf("create-post-session:%d-%s", userID, sessionID)
}

func postLikesKey(postID int64) string {
	return fmt.Sprintf("post-likes:%d", postID)
}

type RedisCacheRepository struct {
	conn *redis.Client
	// postRepo domain.PostRepository
}

func NewRedisCacheRepository(
	conn *redis.Client,
	// postRepo domain.PostRepository,
) *RedisCacheRepository {
	return &RedisCacheRepository{
		conn: conn,
		// postRepo: postRepo,
	}
}

func (rc *RedisCacheRepository) SetUploadPostSession(ctx context.Context, userID int64, sessionID string, createPostSession *domain.CreatePostSession) error {
	key := createPostSessionKey(userID, sessionID)

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

func (rc *RedisCacheRepository) GetUploadPostSession(ctx context.Context, userID int64, sessionID string) (domain.CreatePostSession, error) {
	key := createPostSessionKey(userID, sessionID)

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

func (rc *RedisCacheRepository) IncPostLikes(ctx context.Context, postID int64) error {
	key := postLikesKey(postID)
	if _, err := rc.conn.Exists(ctx, key).Result(); err != nil {
		return errs.NewError(errs.NotFound, err, errs.PostLikeCacheNotFound)
	}

	if _, err := rc.conn.Incr(ctx, key).Result(); err != nil {
		return fmt.Errorf("RedisCacheRepository.IncPostLikes: %w", err)
	}
	return nil
}

func (rc *RedisCacheRepository) SetPostLikesIfNotExists(ctx context.Context, postID int64, likeCount int) error {
	key := postLikesKey(postID)

	if _, err := rc.conn.SetNX(ctx, key, likeCount, 0).Result(); err != nil {
		return fmt.Errorf("RedisCacheRepository.IncPostLikes: %w", err)
	}
	return nil
}

func (rc *RedisCacheRepository) DecPostLikes(ctx context.Context, postID int64) error {
	key := postLikesKey(postID)
	if _, err := rc.conn.Exists(ctx, key).Result(); err != nil {
		return errs.NewError(errs.NotFound, err, errs.PostLikeCacheNotFound)
	}

	if _, err := rc.conn.Decr(ctx, key).Result(); err != nil {
		return fmt.Errorf("RedisCacheRepository.IncPostLikes: %w", err)
	}
	return nil
}

func (rc *RedisCacheRepository) GetLikeCount(ctx context.Context, postIDs ...int64) (map[int64]int64, error) {
	len := len(postIDs)
	if len == 0{
		return map[int64]int64{}, nil
	}
	var keys []string

	for _, id := range postIDs {
		keys = append(keys, postLikesKey(id))
	}

	vals, err := rc.conn.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("RedisCacheRepository.IncPostLikes: %w", err)
	}

	postLikes := make(map[int64]int64, 0)
	i := 0
	for i < len {
		if vals[i] != nil {
			likeCount, err := strconv.ParseInt(vals[i].(string), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("RedisCacheRepository.IncPostLikes: %w", err)
			}
			postLikes[postIDs[i]] = likeCount
		}
		i++
	}
	return postLikes, nil
}
