package feature

import (
	"context"
	"fmt"
	"log"

	"github.com/htan06/Moments/post/internal/application/ports"
	"github.com/htan06/Moments/post/internal/domain/event"
)

type LikePostCmd struct {
	UserID int64
	PostID int64
}

type LikePostUC struct {
	postCreator ports.PostLiker
	// cacheRepo    ports.CacheRepository
	postProducer ports.PostEventPublisher
}

func NewLikePostUC(
	postCreator ports.PostLiker,
	// cacheRepo ports.CacheRepository,
	postProducer ports.PostEventPublisher,
) *LikePostUC {
	return &LikePostUC{
		postCreator: postCreator,
		// cacheRepo:    cacheRepo,
		postProducer: postProducer,
	}
}

func (l *LikePostUC) Execute(ctx context.Context, cmd LikePostCmd) error {
	if err := l.postCreator.CreateLikePost(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("LikePostUC.Execute: %w", err)
	}

	go func() {
		err := l.postProducer.SendPostLiked(context.Background(), event.NewPostLiked(cmd.PostID, cmd.UserID))
		if err != nil {
			log.Println(err.Error())
		}
	}()

	return nil
}
