package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/htan06/echo-messenger-rest-api/internal/config"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
)

type Content struct {
	Type domain.NodeType `json:"type"`
	Text string          `json:"text"`
}

type CreatePostCmd struct {
	AuthorID    int64
	Contents    []Content
	Visibility  domain.Visibility
	MediaCounts int
}

type CreatePostRes struct {
	PostSessionID string
	UploadURLs    []string
}

type CreatePostUC struct {
	// postRepo      domain.PostRepository
	userRepo      domain.UserRepository
	objectStorage domain.ObjectStorage
	cacheRepo     domain.CacheRepository
}

func NewCreatePostUC(
	// postRepo domain.PostRepository,
	userRepo domain.UserRepository,
	objectStorage domain.ObjectStorage,
	cacheRepo domain.CacheRepository,
) *CreatePostUC {
	return &CreatePostUC{
		// postRepo:      postRepo,
		userRepo:      userRepo,
		objectStorage: objectStorage,
		cacheRepo:     cacheRepo,
	}
}

func (cp *CreatePostUC) Excute(ctx context.Context, cmd CreatePostCmd) (CreatePostRes, error) {
	postPending := domain.PostPending{}

	postPending.AuthorID = cmd.AuthorID
	postPending.Visibility = cmd.Visibility

	for _, n := range cmd.Contents {
		switch n.Type {
		case domain.Text:
			postPending.Contents = append(postPending.Contents, domain.Node{
				Type:  n.Type,
				Value: n.Text,
			})
			break

		case domain.Mention:
			userID, err := cp.userRepo.GetIDByUsername(ctx, n.Text[1:])
			if err != nil {
				postPending.Contents = append(postPending.Contents, domain.Node{
					Type:  domain.Text,
					Value: n.Text,
				})
			} else {
				postPending.Contents = append(postPending.Contents, domain.Node{
					Type:  domain.Mention,
					Value: userID,
				})
			}
			break

		case domain.Hashtag:
			_, val, _ := strings.Cut(n.Text, "#")
			postPending.Contents = append(postPending.Contents, domain.Node{
				Type:  domain.Hashtag,
				Value: val,
			})
		}
	}

	var uploadURLs []string

	for i := 0; i < cmd.MediaCounts; i++ {
		randID, err := uuid.NewRandom()
		if err != nil {
			return CreatePostRes{}, fmt.Errorf("CreatePostUC.Excute: %w", err)
		}

		mediaID := randID.String()

		presignedUploadURL, err := cp.objectStorage.GetPresignedURLUpload(ctx, string(config.TempBucket), mediaID, time.Minute*5)
		if err != nil {
			return CreatePostRes{}, fmt.Errorf("CreatePostUC.Excute: %w", err)
		}

		uploadURLs = append(uploadURLs, presignedUploadURL)
		postPending.MediaIDs = append(postPending.MediaIDs, mediaID)
	}

	randID, err := uuid.NewRandom()
	if err != nil {
		return CreatePostRes{}, fmt.Errorf("CreatePostUC.Excute: %w", err)
	}

	sessionID := fmt.Sprintf("%d:%s", cmd.AuthorID, randID.String())

	if err := cp.cacheRepo.SetPostPending(ctx, sessionID, postPending); err != nil {
		return CreatePostRes{}, fmt.Errorf("CreatePostUC.Excute: %w", err)
	}

	return CreatePostRes{
		PostSessionID: sessionID,
		UploadURLs:    uploadURLs,
	}, nil
}
