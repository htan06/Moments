package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/redis/go-redis/v9"
)

func createPostSessionKey(userID int64, sessionID string) string {
	return fmt.Sprintf("create-post-session:%d-%s", userID, sessionID)
}

type RedisSessionRepository struct {
	conn *redis.Client
	// postRepo domain.PostRepository
}

func NewRedisSessionRepository(
	conn *redis.Client,
	// postRepo domain.PostRepository,
) *RedisSessionRepository {
	return &RedisSessionRepository{
		conn: conn,
		// postRepo: postRepo,
	}
}

func (rc *RedisSessionRepository) SetUploadPostSession(ctx context.Context, createPostSession *domain.CreatePostSession) error {
	key := createPostSessionKey(createPostSession.UserID, createPostSession.SessionID.String())

	data, err := json.Marshal(*createPostSession)
	if err != nil {
		return fmt.Errorf("RedisSessionRepository.SetUploadPostSession: %w", err)
	}

	cmd := rc.conn.Set(ctx, key, data, time.Minute*5)
	if err := cmd.Err(); err != nil {
		return fmt.Errorf("RedisSessionRepository.SetUploadPostSession: %w", err)
	}
	return nil
}

func (rc *RedisSessionRepository) GetUploadPostSession(ctx context.Context, userID int64, sessionID string) (domain.CreatePostSession, error) {
	key := createPostSessionKey(userID, sessionID)

	data, err := rc.conn.Get(ctx, key).Bytes()
	if err != nil {
		return domain.CreatePostSession{}, fmt.Errorf("RedisSessionRepository.GetUploadPostSession: %w", err)
	}

	var uploadPostSession domain.CreatePostSession
	if err := json.Unmarshal(data, &uploadPostSession); err != nil {
		return domain.CreatePostSession{}, fmt.Errorf("RedisSessionRepository.GetUploadPostSession: %w", err)
	}

	return uploadPostSession, nil
}
