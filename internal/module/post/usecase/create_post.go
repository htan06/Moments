package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
)

type CreatePostCmd struct {
	AuthorID int64
	Content  []struct {
		Type domain.NodeType
		Text string
	}
	MediaCounts int
}

type CreatePostUC struct {
	postRepo      domain.PostRepository
	userRepo      domain.UserRepository
	objectStorage domain.ObjectStorage
}

func NewCreatePostUC(
	postRepo domain.PostRepository,
	userRepo domain.UserRepository,
	objectStorage domain.ObjectStorage,
) *CreatePostUC {
	return &CreatePostUC{
		postRepo:      postRepo,
		userRepo:      userRepo,
		objectStorage: objectStorage,
	}
}

func (cp *CreatePostUC) Excute(ctx context.Context, cmd CreatePostCmd) ([]string, error) {
	var post domain.Post

	post.AuthorID = cmd.AuthorID

	for _, n := range cmd.Content {
		switch n.Type {
		case domain.Text:
			post.Content = append(post.Content, domain.Node{
				Type:  n.Type,
				Value: n.Text,
			})
			break

		case domain.Mention:
			userID, err := cp.userRepo.GetIDByUsername(ctx, n.Text[1:])
			if err != nil {
				post.Content = append(post.Content, domain.Node{
					Type:  domain.Text,
					Value: n.Text,
				})
			} else {
				post.Content = append(post.Content, domain.Node{
					Type:  domain.Mention,
					Value: userID,
				})
			}
			break

		case domain.Hashtag:
			post.Content = append(post.Content, domain.Node{
				Type:  domain.Hashtag,
				Value: n.Text,
			})
		}
	}

	var listURLS []string

	for i := 0; i < cmd.MediaCounts; i++ {
		randID, err := uuid.NewRandom()
		if err != nil {
			return nil, fmt.Errorf("CreatePostUC.Excute: %w", err)
		}

		mediaID := fmt.Sprintf("/posts/%s", randID.String())

		presignedUploadURL, err := cp.objectStorage.GetPresignedURLUpload(ctx, "medias", mediaID, time.Minute*5)
		if err != nil {
			return nil, fmt.Errorf("CreatePostUC.Excute: %w", err)
		}

		listURLS = append(listURLS, presignedUploadURL)
		post.Medias = append(post.Medias, domain.Media{
			Type: "",
		})
	}
	return nil, nil
}
