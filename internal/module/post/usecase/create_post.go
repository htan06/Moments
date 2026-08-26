package usecase

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

type UploadPostCmd struct {
	UpLoadSessionID string
	AuthorID        int64
	Contents        []domain.Content
	Visibility      domain.Visibility
	AspectRatio     string
	MediaIDs        uuid.UUIDs
}

type CreatePostSessionCmd struct {
	UserID     int64
	MediaCount int
}

type MediaUpload struct {
	MediaID   string `json:"media_id"`
	UploadURL string `json:"upload_url"`
}

type RequestUploadURLs struct {
	SessionID  uuid.UUID
	AuthorID   int64
	MediaCount int
}

type CreatePostSessionRes struct {
	SessionID    string
	MediaUploads []MediaUpload
}

type UploadPostRes struct {
	PostSessionID string
	MediaUploads  []MediaUpload
}

type CreatePostUC struct {
	postRepo      domain.PostRepository
	userRepo      domain.UserRepository
	objectStorage domain.ObjectStorage
	cacheRepo     domain.CacheRepository
	imgProcessor  domain.ProcessImg
	postProducer  domain.PostProducer
}

func NewCreatePostUC(
	postRepo domain.PostRepository,
	userRepo domain.UserRepository,
	objectStorage domain.ObjectStorage,
	cacheRepo domain.CacheRepository,
	imgProcessor domain.ProcessImg,
	postProducer domain.PostProducer,
) *CreatePostUC {
	return &CreatePostUC{
		postRepo:      postRepo,
		userRepo:      userRepo,
		objectStorage: objectStorage,
		cacheRepo:     cacheRepo,
		imgProcessor:  imgProcessor,
		postProducer:  postProducer,
	}
}

func (cp *CreatePostUC) ExecuteRequestUploadURLs(ctx context.Context, cmd RequestUploadURLs) ([]MediaUpload, error) {
	createPostSession, err := cp.cacheRepo.GetUploadPostSession(ctx, cmd.AuthorID, cmd.SessionID.String())
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	mediaIDs, err := createPostSession.RequestMediaIDs(cmd.MediaCount)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	var mediaUploads []MediaUpload
	for _, m := range mediaIDs {
		url, err := cp.objectStorage.GetPresignedURLUpload(ctx, string(config.TempBucket), m.String(), time.Minute*30)
		if err != nil {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}
		mediaUploads = append(mediaUploads, MediaUpload{MediaID: m.String(), UploadURL: url})
	}

	if err := cp.cacheRepo.SetUploadPostSession(ctx, cmd.AuthorID, cmd.SessionID.String(), &createPostSession); err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}
	return mediaUploads, nil
}

func (cp *CreatePostUC) ExecuteCreatePostSession(ctx context.Context, cmd CreatePostSessionCmd) (CreatePostSessionRes, error) {
	createPostSession, err := domain.NewCreatePostSession(cmd.UserID, cmd.MediaCount)
	if err != nil {
		return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
	}

	var mediaUploads []MediaUpload

	for k, _ := range createPostSession.MediaIDs {
		url, err := cp.objectStorage.GetPresignedURLUpload(ctx, string(config.TempBucket), k.String(), time.Minute*30)
		if err != nil {
			return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
		}
		mediaUploads = append(mediaUploads, MediaUpload{MediaID: k.String(), UploadURL: url})

	}

	if err := cp.cacheRepo.SetUploadPostSession(ctx, createPostSession.UserID, createPostSession.SessionID.String(), createPostSession); err != nil {
		return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
	}

	return CreatePostSessionRes{
		SessionID:    createPostSession.SessionID.String(),
		MediaUploads: mediaUploads,
	}, nil
}

func (cp *CreatePostUC) ExecuteUploadPost(ctx context.Context, cmd UploadPostCmd) (*int64, error) {
	createPostSession, err := cp.cacheRepo.GetUploadPostSession(ctx, cmd.AuthorID, cmd.UpLoadSessionID)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	post, err := domain.NewPost(
		cmd.AuthorID,
		cmd.Visibility,
		cmd.Contents,
		cmd.AspectRatio,
	)

	if err != nil {
		return nil, err
	}

	cp.parseContent(ctx, post)

	var firstMedia uuid.UUID
	for k, _ := range createPostSession.MediaIDs {
		firstMedia = k
		break
	}

	thumbnailID, err := cp.createPostThumbnail(ctx, firstMedia)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}
	post.SetThumbnailID(*thumbnailID)

	aw, ah := post.AspectRatio().Demensions()
	mediaW := 1080
	mediaH := (ah * 1080) / aw

	var medias []domain.Media
	for index, mediaID := range cmd.MediaIDs {
		if _, ok := createPostSession.MediaIDs[mediaID]; !ok {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}
		src, err := cp.objectStorage.GetObject(ctx, string(config.TempBucket), mediaID.String())

		buf := make([]byte, 512)
		if _, err := io.ReadFull(src, buf); err != nil {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}
		MIMEType := http.DetectContentType(buf)
		parts := strings.Split(MIMEType, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}
		contentType := parts[0]

		switch contentType {
		case "image":
			newImg, size, err := cp.processImage(ctx, src, mediaW, mediaH)
			if err != nil {
				return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
			}
			if err := cp.objectStorage.PutObject(ctx, string(config.PostBucket), mediaID.String(), newImg, int64(size)); err != nil {
				return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
			}
		case "video":
				
		}

		medias = append(medias,
			domain.Media{
				MediaID:      mediaID,
				Type:         domain.Image,
				DisplayOrder: index + 1,
				Width:        mediaW,
				Height:       mediaH,
			},
		)

	}

	post.SetMedias(medias)
	postId, err := cp.postRepo.CreatePost(ctx, *post)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	if err := cp.postProducer.Send(ctx, domain.PostEvent{
		AuthorID: cmd.AuthorID,
		PostID:   *postId,
		Type:     domain.Created,
	}); err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	return postId, nil
}

func (cp *CreatePostUC) createPostThumbnail(ctx context.Context, mediaID uuid.UUID) (*string, error) {

	media, err := cp.objectStorage.GetObject(ctx, string(config.TempBucket), mediaID.String())
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

func (cp *CreatePostUC) parseContent(ctx context.Context, p *domain.Post) {
	for i := range p.Contents() {
		switch p.Contents()[i].Type {
		case domain.Mention:
			_, username, _ := strings.Cut(p.Contents()[i].Text, "@")
			userID, err := cp.userRepo.GetIDByUsername(ctx, username)
			if err != nil {
				p.Contents()[i].Type = domain.Text
			} else {
				p.AppendMention(*userID)
				p.Contents()[i].Text = strconv.Itoa(int(*userID))
			}

		case domain.Hashtag:
			_, hashtag, _ := strings.Cut(p.Contents()[i].Text, "#")
			p.Appendhashtag(hashtag)
		}
	}
}

func (cp *CreatePostUC) processImage(ctx context.Context, imageSrc io.Reader, width int, height int) (io.Reader, int, error) {
	media, size, err := cp.imgProcessor.Resize(imageSrc, width, height)
	if err != nil {
		return nil, 0, fmt.Errorf("CreatePostUC.processImage: %w", err)
	}

	return media, size, err
}

func (cp *CreatePostUC) processVideo(ctx context.Context, videoSrc io.Reader, width int, height int) (io.Reader, int, error) {
	media, size, err := cp.imgProcessor.Resize(videoSrc, width, height)
	if err != nil {
		return nil, 0, fmt.Errorf("CreatePostUC.processImage: %w", err)
	}

	return media, size, err
}
