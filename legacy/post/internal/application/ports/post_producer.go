package ports

import (
	"context"

	"github.com/htan06/Moments/post/internal/domain/event"
	"github.com/htan06/Moments/post/internal/domain/job"
)

type PostEventPublisher interface {
	SendPostCreated(ctx context.Context, e event.PostCreated) error
	SendPostDeleted(ctx context.Context, e event.PostDeleted) error
	SendPostLiked(ctx context.Context, e event.PostLiked) error
	SendPostUnliked(ctx context.Context, e event.PostUnliked) error
}

type MediaJobPublisher interface {
	SendProcessMediaJob(ctx context.Context, job job.ProcessMediaJob) error
}
