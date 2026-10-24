package feature

import (
	"context"
	"fmt"
	"log"

	"github.com/htan06/Moments/post/internal/application/domain/event"
	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type LikePostCmd struct {
	UserID int64
	PostID int64
}

type LikePostUC struct {
	postRepo outbound.PostRepository
	// cacheRepo    outbound.CacheRepository
	postProducer outbound.PostProducer
}

func NewLikePostUC(
	postRepo outbound.PostRepository,
	// cacheRepo outbound.CacheRepository,
	postProducer outbound.PostProducer,
) *LikePostUC {
	return &LikePostUC{
		postRepo: postRepo,
		// cacheRepo:    cacheRepo,
		postProducer: postProducer,
	}
}

func (l *LikePostUC) Execute(ctx context.Context, cmd LikePostCmd) error {
	if err := l.postRepo.CreateLikePost(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("LikePostUC.Execute: %w", err)
	}

	go func() {
		err := l.postProducer.SendInteractionEvent(context.Background(), event.InteractionEvent{
			UserID:          cmd.UserID,
			PostID:          cmd.PostID,
			Action:          event.Created,
			TypeInteraction: event.LikeInteraction,
		})
		if err != nil {
			log.Println(err.Error())
		}
	}()

	return nil
}
