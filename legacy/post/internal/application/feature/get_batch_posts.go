package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/internal/application/domain/readmodel"
	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type GetBatchPostsQry struct {
	ViewerID int64
	Ids      []int64 `json:"ids"`
}

type GetBatchPostsResp struct {
	Posts []readmodel.Post `json:"posts"`
}

type GetBatchPostsUC struct {
	postRepo outbound.PostRepository
}

func NewGetBatchPostsUC(postRepo outbound.PostRepository) *GetBatchPostsUC {
	return &GetBatchPostsUC{
		postRepo: postRepo,
	}
}

func (uc *GetBatchPostsUC) Execute(ctx context.Context, qry GetBatchPostsQry) (GetBatchPostsResp, error) {
	posts, err := uc.postRepo.GetBatchPosts(ctx, qry.ViewerID, qry.Ids)
	if err != nil {
		return GetBatchPostsResp{}, fmt.Errorf("GetBatchPostsUC.Execute: %w", err)
	}

	return GetBatchPostsResp{Posts: posts}, err
}
