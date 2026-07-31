package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
)

type DeletePostCmd struct {
	UserID int64
	PostID int64
}

type DeletePostUC struct {
	postRepo domain.PostRepository
}

func NewDeletePostUC(
	postRepo domain.PostRepository,
) *DeletePostUC {
	return &DeletePostUC{
		postRepo: postRepo,
	}
}

func (dp *DeletePostUC) Execute(ctx context.Context, cmd DeletePostCmd) error {
	if err := dp.postRepo.DeletePostByUserIDAndPostID(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("DeletePostUC.Execute: %w", err)
	}
	return nil
}