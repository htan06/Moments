package feature

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/post/internal/application/ports"
	"github.com/htan06/Moments/post/internal/application/readmodel"
	"github.com/htan06/Moments/post/internal/config"
)

type GetFollowingFeedQry struct {
	ViewerID int64
	Cursor   int64 `form:"cursor"`
	Size     int   `form:"size"`
}

type GetFollowingFeedResp struct {
	Posts         []readmodel.Post `json:"posts"`
	CurrentCursor int64            `json:"current_cursor"`
}

type GetFollowingFeedUC struct {
	postReader ports.FollowingFeedReader
}

func NewGetFollowingFeedUC(
	postReader ports.FollowingFeedReader,
) *GetFollowingFeedUC {
	return &GetFollowingFeedUC{
		postReader: postReader,
	}
}

func (uc *GetFollowingFeedUC) Execute(ctx context.Context, qry GetFollowingFeedQry) (GetFollowingFeedResp, error) {
	posts, err := uc.postReader.GetPostsByFollowing(ctx, qry.ViewerID, qry.Cursor/1000000.0, qry.Size)
	if err != nil {
		return GetFollowingFeedResp{}, fmt.Errorf("GetPostUC.Excute: %w", err)
	}

	lenOfPost := len(posts)
	ids := make([]int64, 0)
	for i := range posts {
		readmodel.EnrichContent(posts[i].Content)

		posts[i].ThumbnailID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostsStorage, posts[i].ThumbnailID)

		lenOfMedias := len(posts[i].Medias)
		for j := 0; j < lenOfMedias; j++ {
			posts[i].Medias[j].MediaID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostsStorage, posts[i].Medias[j].MediaID)
		}
		ids = append(ids, posts[i].ID)
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
