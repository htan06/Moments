package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/post/internal/domain"

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
	postRepo       domain.PostRepository
	userRepo       domain.UserRepository
	objectStorage  domain.ObjectStorage
	cacheRepo      domain.CacheRepository
	imgProcessor   domain.ProcessImg
	videoProcessor domain.ProcessVideo
	postProducer   domain.PostProducer
}

func NewCreatePostUC(
	postRepo domain.PostRepository,
	userRepo domain.UserRepository,
	objectStorage domain.ObjectStorage,
	cacheRepo domain.CacheRepository,
	imgProcessor domain.ProcessImg,
	videoProcessor domain.ProcessVideo,
	postProducer domain.PostProducer,
) *CreatePostUC {
	return &CreatePostUC{
		postRepo:       postRepo,
		userRepo:       userRepo,
		objectStorage:  objectStorage,
		cacheRepo:      cacheRepo,
		imgProcessor:   imgProcessor,
		videoProcessor: videoProcessor,
		postProducer:   postProducer,
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

	firstMedia := cmd.MediaIDs[0]

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
		if err != nil {
			return nil, err
		}

		buf := make([]byte, 262)
		if _, err := io.ReadFull(src, buf); err != nil {
			return nil, err
		}

		if _, err := src.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}

		if filetype.IsImage(buf) {
			newImg, size, err := cp.processImage(ctx, src, mediaW, mediaH)
			if err != nil {
				return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
			}
			if err := cp.objectStorage.PutObject(ctx, string(config.PostBucket), mediaID.String(), newImg, int64(size)); err != nil {
				return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
			}

			medias = append(medias,
				domain.Media{
					MediaID:      mediaID.String(),
					Type:         domain.Image,
					DisplayOrder: index + 1,
					Width:        mediaW,
					Height:       mediaH,
				},
			)
		} else if filetype.IsVideo(buf) {
			url, err := cp.objectStorage.GetPresignedURLDownload(ctx, string(config.TempBucket), mediaID.String(), time.Minute*10)
			if err != nil {
				return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
			}

			if err := cp.videoProcessor.HLS(url, mediaID.String()); err != nil {
				return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
			}

			medias = append(medias,
				domain.Media{
					MediaID:      fmt.Sprintf("%s/index.m3u8", mediaID),
					Type:         domain.Video,
					DisplayOrder: index + 1,
					Width:        mediaW,
					Height:       mediaH,
				},
			)

			folderLocation := fmt.Sprintf("./internal/module/post/tmp/%s", mediaID)
			entries, err := os.ReadDir(folderLocation)
			if err != nil {
				return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
			}

			var contentType string

			for _, entry := range entries {
				objName := fmt.Sprintf("%s/%s", mediaID, entry.Name())
				objPath := fmt.Sprintf("%s/%s", folderLocation, entry.Name())

				switch {
				case strings.HasSuffix(entry.Name(), ".m3u8"):
					contentType = "application/vnd.apple.mpegurl"
				case strings.HasSuffix(entry.Name(), ".ts"):
					contentType = "video/mp2t"
				default:
					contentType = "application/octet-stream"
				}

				if err := cp.objectStorage.FPutObject(ctx, string(config.PostBucket), objName, objPath, contentType); err != nil {
					return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
				}
			}
			if err := os.RemoveAll(fmt.Sprintf("./internal/module/post/tmp/%s", mediaID)); err != nil {
				return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
			}
		}
	}

	post.SetMedias(medias)
	postId, err := cp.postRepo.CreatePost(ctx, *post)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	if err := cp.postProducer.SendPostEvent(ctx, domain.PostEvent{
		AuthorID: cmd.AuthorID,
		PostID:   *postId,
		Action:   domain.Created,
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
			p.Contents()[i].Text = hashtag
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
