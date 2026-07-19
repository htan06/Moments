package usecase

import (
	"context"
	"fmt"

	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
)

type GetPostUC struct {
	postRepo domain.PostRepository
}

func NewGetPostUC(
	postRepo domain.PostRepository,
) *GetPostUC {
	return &GetPostUC{
		postRepo: postRepo,
	}
}

func (gp *GetPostUC) Excute(ctx context.Context, postID int64) (domain.PostReadModel, error) {
	post, err := gp.postRepo.GetPost(ctx, postID)
	if err != nil {
		return domain.PostReadModel{}, fmt.Errorf("GetPostUC.Excute: %w", err)
	}

	if post.AuthorAvatarThumbnailID != nil {
		*post.AuthorAvatarThumbnailID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *post.AuthorAvatarThumbnailID)
	}

	for i := range post.Medias {
		post.Medias[i].MediaID = fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.PostBucket, post.Medias[i].MediaID)
	}
	return post, nil
}
