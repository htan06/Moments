package feature

import (
	"context"
	"fmt"

	// "github.com/htan06/Moments/post/internal/adapter/outbound/config"

	"github.com/htan06/Moments/post/internal/application/domain/readmodel"
	"github.com/htan06/Moments/post/internal/config"
	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type Direction string

const (
	Before Direction = "BEFORE"
	After  Direction = "AFTER"
)

type GetPostsByCursorQry struct {
	ViewerID int64
	AuthorID int64 `form:"author_id"`
	Cursor   int64 `form:"cursor"`
	Size     int   `form:"size"`
}

type GetPostsByCursorResp struct {
	IsAuthor      bool                    `json:"is_author"`
	Posts         []readmodel.PostSummary `json:"posts"`
	CurrentCursor int64                   `json:"current_cursor"`
}

type GetPostsByCursorUC struct {
	postRepo    outbound.PostRepository
	counterRepo outbound.CounterRepository
}

func NewGetPostsByCursorUC(
	postRepo outbound.PostRepository,
	counterRepo outbound.CounterRepository,
) *GetPostsByCursorUC {
	return &GetPostsByCursorUC{
		postRepo:    postRepo,
		counterRepo: counterRepo,
	}
}

func (gp *GetPostsByCursorUC) Execute(ctx context.Context, qry GetPostsByCursorQry) (GetPostsByCursorResp, error) {
	posts, err := gp.postRepo.GetPostsByAuthorID(ctx, qry.ViewerID, qry.AuthorID, qry.Cursor/1000000.0, qry.Size)
	if err != nil {
		return GetPostsByCursorResp{}, fmt.Errorf("GetPostUC.Excute: %w", err)
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

	postLike, err := gp.counterRepo.GetLikeCount(ctx, ids...)
	if err != nil {
		return GetPostsByCursorResp{}, fmt.Errorf("GetPostUC.Excute: %w", err)
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

	return GetPostsByCursorResp{
		IsAuthor:      qry.ViewerID == qry.AuthorID,
		Posts:         posts,
		CurrentCursor: cursor,
	}, nil
}
