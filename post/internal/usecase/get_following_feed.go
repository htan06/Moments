package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/config"
	"github.com/htan06/Moments/post/internal/domain"
)

type GetFollowingFeedQry struct {
	ViewerID int64
	Cursor   int64 `form:"cursor"`
	Size     int   `form:"size"`
}

type GetFollowingFeedResp struct {
	Posts         []domain.PostAuth `json:"posts"`
	CurrentCursor int64             `json:"current_cursor"`
}

type GetFollowingFeedUC struct {
	postRepo    domain.PostRepository
	counterRepo domain.CounterRepository
}

func NewGetFollowingFeedUC(
	postRepo domain.PostRepository,
	counterRepo domain.CounterRepository,
) *GetFollowingFeedUC {
	return &GetFollowingFeedUC{
		postRepo:    postRepo,
		counterRepo: counterRepo,
	}
}

func (uc *GetFollowingFeedUC) Execute(ctx context.Context, qry GetFollowingFeedQry) (GetFollowingFeedResp, error) {
	posts, err := uc.postRepo.GetPostsByFollowing(ctx, qry.ViewerID, qry.Cursor/1000000.0, qry.Size)
	if err != nil {
		return GetFollowingFeedResp{}, fmt.Errorf("GetPostUC.Excute: %w", err)
	}

	lenOfPost := len(posts)
	ids := make([]int64, 0)
	for i := range posts {
		domain.EnrichContent(posts[i].Content)

		posts[i].ThumbnailID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostBucket, posts[i].ThumbnailID)

		lenOfMedias := len(posts[i].Medias)
		for j := 0; j < lenOfMedias; j++ {
			posts[i].Medias[j].MediaID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostBucket, posts[i].Medias[j].MediaID)
		}
		ids = append(ids, posts[i].ID)
	}

	postLike, err := uc.counterRepo.GetLikeCount(ctx, ids...)
	if err != nil {
		return GetFollowingFeedResp{}, fmt.Errorf("GetPostUC.Excute: %w", err)
	}

	for i := range posts {
		realtimeLikeCount, ok := postLike[posts[i].ID]
		if ok {
			posts[i].LikeCount = int(realtimeLikeCount)
		}
	}

	var cursor int64
	if lenOfPost != 0 {
		cursor = posts[lenOfPost-1].CreatedAt.UnixMicro()
	} else {
		cursor = 0
	}

	return GetFollowingFeedResp{
		Posts:         posts,
		CurrentCursor: cursor,
	}, nil
}
