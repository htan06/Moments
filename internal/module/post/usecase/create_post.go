package usecase

import (
	"context"
	"fmt"
	"image"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/module/post/domain"

	_ "image/gif"  // Registers GIF decoder
	_ "image/jpeg" // Registers JPEG decoder
	_ "image/png"  // Registers PNG decoder

	// For WebP or BMP, use golang.org/x/image
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

type Content struct {
	Type domain.NodeType `json:"type"`
	Text string          `json:"text"`
}

type UploadPostCmd struct {
	UpLoadSessionID string
	AuthorID        int64
	Contents        []Content
	Visibility      domain.Visibility
	AspectRatio     domain.AspectRatio
}

type CreatePostSessionCmd struct {
	UserID     int64
	MediaCount int
}

type CreatePostSessionRes struct {
	SessionID     string
	PresignedURLs []string
}

type UploadPostRes struct {
	PostSessionID string
	UploadURLs    []string
}

type CreatePostUC struct {
	postRepo      domain.PostRepository
	userRepo      domain.UserRepository
	objectStorage domain.ObjectStorage
	cacheRepo     domain.CacheRepository
	imgProcessor  domain.ProcessImg
}

func NewCreatePostUC(
	postRepo domain.PostRepository,
	userRepo domain.UserRepository,
	objectStorage domain.ObjectStorage,
	cacheRepo domain.CacheRepository,
	imgProcessor domain.ProcessImg,
) *CreatePostUC {
	return &CreatePostUC{
		postRepo:      postRepo,
		userRepo:      userRepo,
		objectStorage: objectStorage,
		cacheRepo:     cacheRepo,
		imgProcessor:  imgProcessor,
	}
}

func (cp *CreatePostUC) ExecuteCreatePostSession(ctx context.Context, cmd CreatePostSessionCmd) (CreatePostSessionRes, error) {
	var uploadPostSession domain.CreatePostSession
	var presignedURLs []string

	for i := 0; i < cmd.MediaCount; i++ {
		randID, err := uuid.NewRandom()
		if err != nil {
			return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
		}

		mediaID := randID.String()

		presignedUploadURL, err := cp.objectStorage.GetPresignedURLUpload(ctx, string(config.TempBucket), mediaID, time.Minute*5)
		if err != nil {
			return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
		}

		presignedURLs = append(presignedURLs, presignedUploadURL)
		uploadPostSession.MediaIDs = append(uploadPostSession.MediaIDs, mediaID)
	}

	sessionID, err := uuid.NewRandom()
	if err != nil {
		return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
	}

	key := fmt.Sprintf("%d:%s", cmd.UserID, sessionID)
	if err := cp.cacheRepo.SetUploadPostSession(ctx, key, uploadPostSession); err != nil {
		return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
	}

	return CreatePostSessionRes{
		SessionID:     sessionID.String(),
		PresignedURLs: presignedURLs,
	}, nil
}

func (cp *CreatePostUC) ExecuteUploadPost(ctx context.Context, cmd UploadPostCmd) (*int64, error) {
	key := fmt.Sprintf("%d:%s", cmd.AuthorID, cmd.UpLoadSessionID)

	uploadPostSession, err := cp.cacheRepo.GetUploadPostSession(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	post := domain.Post{
		AuthorID:    cmd.AuthorID,
		Visibility:  cmd.Visibility,
		AspectRatio: cmd.AspectRatio,
		MediaCount:  len(uploadPostSession.MediaIDs),
	}

	for _, n := range cmd.Contents {
		switch n.Type {
		case domain.Text:
			post.Contents = append(post.Contents, domain.Content{
				Type:  n.Type,
				Value: n.Text,
			})

		case domain.Mention:
			userID, err := cp.userRepo.GetIDByUsername(ctx, n.Text[1:])
			if err != nil {
				post.Contents = append(post.Contents, domain.Content{
					Type:  domain.Text,
					Value: n.Text,
				})
			} else {
				uID := *userID
				post.Mentions = append(post.Mentions, uID)
				post.Contents = append(post.Contents, domain.Content{
					Type:  domain.Mention,
					Value: strconv.Itoa(int(uID)),
				})
			}

		case domain.Hashtag:
			_, val, _ := strings.Cut(n.Text, "#")	
			post.Contents = append(post.Contents, domain.Content{
				Type:  domain.Hashtag,
				Value: val,
			})
		}
	}

	firstMediaID := uploadPostSession.MediaIDs[0]

	thumbnailID, err := cp.createPostThumbnail(ctx, domain.Ratio3_4, firstMediaID)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	post.ThumbnailID = *thumbnailID

	for index, mediaID := range uploadPostSession.MediaIDs {
		src, err := cp.objectStorage.GetObject(ctx, string(config.TempBucket), mediaID)

		if err != nil {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}

		config, _, err := image.DecodeConfig(src)
		if err != nil {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}

		post.Medias = append(post.Medias,
			domain.Media{
				MediaID:      mediaID,
				Type:         domain.Image,
				DisplayOrder: index + 1,
				Width:        config.Width,
				Height:       config.Height,
			},
		)

	}

	postId, err := cp.postRepo.CreatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	for _, mediaID := range uploadPostSession.MediaIDs {
		if err := cp.objectStorage.PromotePostImage(ctx, mediaID); err != nil {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}
	}

	return postId, nil
}

func (cp *CreatePostUC) createPostThumbnail(ctx context.Context, postAspectratio domain.AspectRatio, mediaID string) (*string, error) {

	media, err := cp.objectStorage.GetObject(ctx, string(config.TempBucket), mediaID)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.createPostThumbnail: %w", err)
	}

	thumbnail, size, err := cp.imgProcessor.Resize(media, 300, 400)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.createPostThumbnail: %w", err)
	}

	randID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.createPostThumbnail: %w", err)
	}

	thubmnailID := randID.String()
	if err := cp.objectStorage.PutObject(ctx, string(config.PostBucket), thubmnailID, thumbnail, int64(size)); err != nil {
		return nil, fmt.Errorf("CreatePostUC.createPostThumbnail: %w", err)
	}
	return &thubmnailID, nil
}
