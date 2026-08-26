package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/module/post/domain"
)

type DeletePostCmd struct {
	UserID int64
	PostID int64
}

type DeletePostUC struct {
	postRepo     domain.PostRepository
	postProducer domain.PostProducer
}

func NewDeletePostUC(
	postRepo domain.PostRepository,
	postProducer domain.PostProducer,
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

	if err := dp.postProducer.Send(ctx, domain.PostEvent{
		AuthorID: cmd.UserID,
		PostID:   cmd.PostID,
		Type:     domain.Deleted,
	}); err != nil {
		return fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	return nil
}
