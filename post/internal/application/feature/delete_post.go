package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/internal/application/domain/event"
	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type DeletePostCmd struct {
	UserID int64
	PostID int64
}

type DeletePostUC struct {
	postRepo     outbound.PostRepository
	postProducer outbound.PostProducer
}

func NewDeletePostUC(
	postRepo outbound.PostRepository,
	postProducer outbound.PostProducer,
) *DeletePostUC {
	return &DeletePostUC{
		postRepo:     postRepo,
		postProducer: postProducer,
	}
}

func (dp *DeletePostUC) Execute(ctx context.Context, cmd DeletePostCmd) error {
	if err := dp.postRepo.DeletePostByUserIDAndPostID(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("DeletePostUC.Execute: %w", err)
	}

	if err := dp.postProducer.SendPostEvent(ctx, event.PostEvent{
		AuthorID: cmd.UserID,
		PostID:   cmd.PostID,
		Action:   event.Deleted,
	}); err != nil {
		return fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	return nil
}
