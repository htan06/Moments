package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type DeleteRepostCmd struct {
	UserID   int64
	RepostID int64
}

type DeleteRepostUC struct {
	postRepo outbound.PostRepository
}

func NewDeleteRepostUC(postRepo outbound.PostRepository) *DeleteRepostUC {
	return &DeleteRepostUC{
		postRepo: postRepo,
	}
}

func (uc *DeleteRepostUC) Execute(ctx context.Context, cmd DeleteRepostCmd) error {
	if err := uc.postRepo.DeleteRepost(ctx, cmd.UserID, cmd.RepostID); err != nil {
		return fmt.Errorf("DeleteRepostUC.Excecute: %w", err)
	}

	return nil
}
