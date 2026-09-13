package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/htan06/Moments/post/internal/domain"
)

type UnlikePostCmd struct {
	UserID int64
	PostID int64
}

type UnlikePostUC struct {
	postRepo     domain.PostRepository
	postProducer domain.PostProducer
}

func NewUnlikePostUC(
	postRepo domain.PostRepository,
	postProducer domain.PostProducer,
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
		err := l.postProducer.SendInteractionEvent(context.Background(), domain.InteractionEvent{
			UserID:          cmd.UserID,
			PostID:          cmd.PostID,
			Action:          domain.Deleted,
			TypeInteraction: domain.LikeInteraction,
		})
		if err != nil {
			log.Println(err.Error())
		}
	}()

	return nil
}
