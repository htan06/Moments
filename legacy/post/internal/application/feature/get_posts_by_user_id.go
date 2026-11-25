package feature

import (
	"context"
	"fmt"

	// "github.com/htan06/Moments/post/internal/adapter/outbound/config"

	"github.com/htan06/Moments/post/internal/application/ports"
	"github.com/htan06/Moments/post/internal/application/readmodel"
	"github.com/htan06/Moments/post/internal/config"
)

type Direction string

const (
	Before Direction = "BEFORE"
	After  Direction = "AFTER"
)

type GetPostsByCursorUC struct {
	postReader ports.PostByAuthorReader
}

func NewGetPostsByCursorUC(
	postReader ports.PostByAuthorReader,
) *GetPostsByCursorUC {
	return &GetPostsByCursorUC{
		postReader: postReader,
	}
}

type GetPostsByCursorQry struct {
	ViewerID int64
	AuthorID int64 `form:"author_id"`
	Cursor   int64 `form:"cursor"`
	Size     int   `form:"size"`
}

type GetPostsByCursorResp struct {
	IsAuthor      bool             `json:"is_author"`
	Posts         []readmodel.Post `json:"posts"`
	CurrentCursor int64            `json:"current_cursor"`
}

func (gp *GetPostsByCursorUC) Execute(ctx context.Context, qry GetPostsByCursorQry) (GetPostsByCursorResp, error) {
	posts, err := gp.postReader.GetPostsByAuthorID(ctx, qry.ViewerID, qry.AuthorID, qry.Cursor/1000000.0, qry.Size)
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
