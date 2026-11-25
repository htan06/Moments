package redisadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/htan06/Moments/post/internal/domain/entity"
	"github.com/redis/go-redis/v9"
)

func createPostSessionKey(userID int64, sessionID string) string {
	return fmt.Sprintf("create-post-session:%d-%s", userID, sessionID)
}

type RedisSessionRepository struct {
	conn *redis.Client
	// postRepo entity.PostRepository
}

func NewRedisSessionRepository(
	conn *redis.Client,
	// postRepo entity.PostRepository,
) *RedisSessionRepository {
	return &RedisSessionRepository{
		conn: conn,
		// postRepo: postRepo,
	}
}

func (rc *RedisSessionRepository) SetCreatePostSession(ctx context.Context, createPostSession *entity.CreatePostSession) error {
	key := createPostSessionKey(createPostSession.UserID(), createPostSession.SessionID().String())

	data, err := createPostSession.MarshalJSON()
	if err != nil {
		return fmt.Errorf("RedisSessionRepository.SetUploadPostSession: %w", err)
	}

	cmd := rc.conn.Set(ctx, key, data, time.Minute*15)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("RedisSessionRepository.SetUploadPostSession: %w", err)
	}
	return nil
}

func (rc *RedisSessionRepository) GetCreatePostSession(ctx context.Context, userID int64, sessionID string) (entity.CreatePostSession, error) {
	key := createPostSessionKey(userID, sessionID)

	data, err := rc.conn.Get(ctx, key).Bytes()
	if err != nil {
		return entity.CreatePostSession{}, fmt.Errorf("RedisSessionRepository.GetUploadPostSession: %w", err)
	}

	session, err := entity.UnmarshalCreatePostSessionJSON(data)
	if err != nil {
		return entity.CreatePostSession{}, fmt.Errorf("RedisSessionRepository.GetUploadPostSession: %w", err)
	}

	return *session, nil
}
