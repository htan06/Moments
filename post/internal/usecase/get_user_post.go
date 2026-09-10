package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/post/internal/domain"
)

type GetUserPostsUC struct {
	postRepo domain.PostRepository
}

func NewGetUserPostsUC(
	postRepo domain.PostRepository,
) *GetUserPostsUC {
	return &GetUserPostsUC{
		postRepo: postRepo,
	}
}

func (gp *GetUserPostsUC) Excute(ctx context.Context, username string) ([]domain.PostGridItem, error) {
	posts, err := gp.postRepo.GetPostsByUsername(ctx, username)
	if err != nil {
		return []domain.PostGridItem{}, fmt.Errorf("GetPostUC.Excute: %w", err)
	}

	for i := range posts {
		posts[i].ThumbnailID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostBucket, posts[i].ThumbnailID)
	}

	return posts, nil
}
