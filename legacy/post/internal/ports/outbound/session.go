package outbound

import (
	"context"

	"github.com/htan06/Moments/post/internal/application/domain"
)

type SessionRepository interface {
	SetUploadPostSession(ctx context.Context, createPostSession *domain.CreatePostSession) error
	GetUploadPostSession(ctx context.Context, userID int64, sessionID string) (domain.CreatePostSession, error)
}
