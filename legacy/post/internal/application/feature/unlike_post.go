package feature

import (
	"context"
	"fmt"
	"log"

	// "github.com/htan06/Moments/post/internal/application/domain/event"
	"github.com/htan06/Moments/post/internal/application/ports"
	"github.com/htan06/Moments/post/internal/domain/event"
)

type UnlikePostCmd struct {
	UserID int64
	PostID int64
}

type UnlikePostUC struct {
	postDeleter  ports.PostLikeRemover
	postProducer ports.PostEventPublisher
}

func NewUnlikePostUC(
	postDeleter ports.PostLikeRemover,
	postProducer ports.PostEventPublisher,
) *UnlikePostUC {
	return &UnlikePostUC{
		postDeleter:  postDeleter,
		postProducer: postProducer,
	}
}

func (l *UnlikePostUC) Execute(ctx context.Context, cmd UnlikePostCmd) error {
	if err := l.postDeleter.DeleteLikePost(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("UnlikePostUC.Execute: %w", err)
	}

	go func() {
		err := l.postProducer.SendPostUnliked(context.Background(), event.NewPostUnliked(cmd.PostID, cmd.UserID))
		if err != nil {
			log.Println(err.Error())
		}
	}()

	return nil
}
