package feature

import (
	"context"
	"fmt"
	"log"

	"github.com/htan06/Moments/post/internal/application/domain/event"
	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type UnlikePostCmd struct {
	UserID int64
	PostID int64
}

type UnlikePostUC struct {
	postRepo     outbound.PostRepository
	postProducer outbound.PostProducer
}

func NewUnlikePostUC(
	postRepo outbound.PostRepository,
	postProducer outbound.PostProducer,
) *UnlikePostUC {
	return &UnlikePostUC{
		postRepo:     postRepo,
		postProducer: postProducer,
	}
}

func (l *UnlikePostUC) Execute(ctx context.Context, cmd UnlikePostCmd) error {
	if err := l.postRepo.DeleteLikePost(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("UnlikePostUC.Execute: %w", err)
	}

	go func() {
		err := l.postProducer.SendInteractionEvent(context.Background(), event.InteractionEvent{
			UserID:          cmd.UserID,
			PostID:          cmd.PostID,
			Action:          event.Deleted,
			TypeInteraction: event.LikeInteraction,
		})
		if err != nil {
			log.Println(err.Error())
		}
	}()

	return nil
}
