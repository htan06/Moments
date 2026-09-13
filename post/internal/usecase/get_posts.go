package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/config"
	"github.com/htan06/Moments/post/internal/domain"
)

type GetPostsQry struct {
	AuthorID int64
	Cursor   int64
	Size     int
}

type GetPostsResp struct {
	Posts         []domain.PostGridItem `json:"posts"`
	CurrentCursor int64                 `json:"current_cursor"`
}

type GetPostsUC struct {
	postRepo domain.PostRepository
}

func NewGetPostsUC(
	postRepo domain.PostRepository,
) *GetPostsUC {
	return &GetPostsUC{
		postRepo: postRepo,
	}
}

func (gp *GetPostsUC) Execute(ctx context.Context, qry GetPostsQry) (GetPostsResp, error) {
	posts, err := gp.postRepo.GetPostsByAuthorID(ctx, qry.AuthorID, qry.Cursor/1000000.0, qry.Size)
	if err != nil {
		return GetPostsResp{}, fmt.Errorf("GetPostUC.Excute: %w", err)
	}

	lenOfPost := len(posts)

	for i := 0; i < lenOfPost; i++ {
		posts[i].ThumbnailID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostBucket, posts[i].ThumbnailID)
	}

	var cursor int64
	if lenOfPost != 0 {
		cursor = posts[lenOfPost - 1].CreatedAt.UnixMicro()
	} else {
		cursor = 0
	}

	return GetPostsResp{
		Posts:         posts,
		CurrentCursor: cursor,
	}, nil
}
