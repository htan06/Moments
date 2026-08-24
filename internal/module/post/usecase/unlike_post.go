package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/module/post/domain"
)

type UnlikePostCmd struct {
	UserID int64
	PostID int64
}

type UnlikePostUC struct {
	postRepo domain.PostRepository
}

func NewUnlikePostUC(
	postRepo domain.PostRepository,
) *UnlikePostUC {
	return &UnlikePostUC{
		postRepo: postRepo,
	}
}

func (l *UnlikePostUC) Execute(ctx context.Context, cmd UnlikePostCmd) error {
	if err := l.postRepo.DeleteLikePost(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("UnlikePostUC.Execute: %w", err)
	}
	return nil
}
