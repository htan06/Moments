package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/internal/domain"
)

type LikePostCmd struct {
	UserID int64
	PostID int64
}

type LikePostUC struct {
	postRepo     domain.PostRepository
	cacheRepo    domain.CacheRepository
	postProducer domain.PostProducer
}

func NewLikePostUC(
	postRepo domain.PostRepository,
	cacheRepo domain.CacheRepository,
	postProducer domain.PostProducer,
) *LikePostUC {
	return &LikePostUC{
		postRepo:     postRepo,
		cacheRepo:    cacheRepo,
		postProducer: postProducer,
	}
}

func (l *LikePostUC) Execute(ctx context.Context, cmd LikePostCmd) error {
	if err := l.postRepo.CreateLikePost(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("LikePostUC.Execute: %w", err)
	}

	go l.postProducer.SendInteractionEvent(ctx, domain.InteractionEvent{
		UserID:          cmd.UserID,
		PostID:          cmd.PostID,
		Action:          domain.Created,
		TypeInteraction: domain.LikeInteraction,
	})

	return nil
}
