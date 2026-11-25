package ports

import (
	"context"

	"github.com/htan06/Moments/post/internal/domain/entity"
)

type SessionSaver interface {
	SetCreatePostSession(ctx context.Context, createPostSession *entity.CreatePostSession) error
}

type SessionLoader interface {
	GetCreatePostSession(ctx context.Context, userID int64, sessionID string) (entity.CreatePostSession, error)
}

type SessionRepository interface {
	SessionSaver
	SessionLoader
}
