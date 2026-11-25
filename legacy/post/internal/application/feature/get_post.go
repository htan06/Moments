package feature

// import (
// 	"context"
// 	"fmt"

// 	"github.com/htan06/Moments/post/internal/application/ports"
// 	"github.com/htan06/Moments/post/internal/application/readmodel"
// 	"github.com/htan06/Moments/post/internal/config"
// )

// type GetPostUC struct {
// 	postRepo ports.PostRepository
// }

// func NewGetPostUC(
// 	postRepo ports.PostRepository,
// ) *GetPostUC {
// 	return &GetPostUC{
// 		postRepo: postRepo,
// 	}
// }

// func (gp *GetPostUC) Execute(ctx context.Context, postID int64) (readmodel.Post, error) {
// 	post, err := gp.postRepo.GetPost(ctx, postID)
// 	if err != nil {
// 		return readmodel.Post{}, fmt.Errorf("GetPostUC.Excute: %w", err)
// 	}

// 	post.ThumbnailID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostsStorage, post.ThumbnailID)

// 	if post.AuthorAvatarThumbnailID != nil {
// 		*post.AuthorAvatarThumbnailID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, *post.AuthorAvatarThumbnailID)
// 	}

// 	for i := range post.Medias {
// 		post.Medias[i].MediaID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostsStorage, post.Medias[i].MediaID)
// 	}
// 	return post, nil
// }
