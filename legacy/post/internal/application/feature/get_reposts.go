package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/internal/application/domain/readmodel"
	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type GetRepostsUC struct {
	postRepo outbound.PostRepository
}

func NewGetRepostsUC(postRepo outbound.PostRepository) *GetRepostsUC {
	return &GetRepostsUC{
		postRepo: postRepo,
	}
}

func (uc *GetRepostsUC) Execute(ctx context.Context, username string) ([]readmodel.PostSummary, error) {
	posts, err := uc.postRepo.GetRepostsByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("GetRepostsUC.Execute: %w", err)
	}

	return posts, nil
}
