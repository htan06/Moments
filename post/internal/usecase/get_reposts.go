package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/internal/domain"
)

type GetRepostsUC struct {
	postRepo domain.PostRepository
}

func NewGetRepostsUC(postRepo domain.PostRepository) *GetRepostsUC {
	return &GetRepostsUC{
		postRepo: postRepo,
	}
}

func (uc *GetRepostsUC) Execute(ctx context.Context, username string) ([]domain.PostGridItem, error) {
	posts, err := uc.postRepo.GetRepostsByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("GetRepostsUC.Execute: %w", err)
	}

	return posts, nil
}
