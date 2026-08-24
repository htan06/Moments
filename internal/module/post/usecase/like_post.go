package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/module/post/domain"
)

type LikePostCmd struct {
	UserID int64
	PostID int64
}

type LikePostUC struct {
	postRepo  domain.PostRepository
	cacheRepo domain.CacheRepository
}

func NewLikePostUC(
	postRepo domain.PostRepository,
	cacheRepo domain.CacheRepository,
) *LikePostUC {
	return &LikePostUC{
		postRepo:  postRepo,
		cacheRepo: cacheRepo,
	}
}

func (l *LikePostUC) Execute(ctx context.Context, cmd LikePostCmd) error {
	if err := l.postRepo.CreateLikePost(ctx, cmd.UserID, cmd.PostID); err != nil {
		return fmt.Errorf("LikePostUC.Execute: %w", err)
	}

	// l.cacheRepo
	return nil
}
