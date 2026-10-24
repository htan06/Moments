package feature

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"github.com/htan06/Moments/post/internal/application/domain"
	"github.com/htan06/Moments/post/internal/application/domain/entity"
	"github.com/htan06/Moments/post/internal/application/domain/event"
	"github.com/htan06/Moments/post/internal/application/domain/vo"
	"github.com/htan06/Moments/post/internal/application/job"
	"github.com/htan06/Moments/post/internal/config"
	"github.com/htan06/Moments/post/internal/ports/outbound"
)

type UploadPostCmd struct {
	UpLoadSessionID string
	AuthorID        int64
	Contents        []vo.Content
	Visibility      entity.Visibility
	AspectRatio     string
	MediaUploads    []MediaUpload
}

type MediaUpload struct {
	MediaID uuid.UUID        `json:"media_id"`
	Type    entity.MediaType `json:"type"`
}

type CreatePostSessionCmd struct {
	UserID     int64
	MediaCount int
}

type MediaUploadRes struct {
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
	MediaUploads []MediaUploadRes
}

type CreatePostUC struct {
	postRepo     outbound.PostRepository
	userRepo     outbound.UserRepository
	mediaStorage outbound.MediaStorage
	sessionRepo  outbound.SessionRepository
	postProducer outbound.PostProducer
}

func NewCreatePostUC(
	postRepo outbound.PostRepository,
	userRepo outbound.UserRepository,
	mediaStorage outbound.MediaStorage,
	sessionRepo outbound.SessionRepository,
	postProducer outbound.PostProducer,
) *CreatePostUC {
	return &CreatePostUC{
		postRepo:     postRepo,
		userRepo:     userRepo,
		mediaStorage: mediaStorage,
		sessionRepo:  sessionRepo,
		postProducer: postProducer,
	}
}

func (cp *CreatePostUC) ExecuteRequestUploadURLs(ctx context.Context, cmd RequestUploadURLs) ([]MediaUploadRes, error) {
	createPostSession, err := cp.sessionRepo.GetUploadPostSession(ctx, cmd.AuthorID, cmd.SessionID.String())
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	mediaIDs, err := createPostSession.RequestMediaIDs(cmd.MediaCount)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	var mediaUploadRes []MediaUploadRes
	for _, m := range mediaIDs {
		url, err := cp.mediaStorage.GetPresignedURLUpload(ctx, m.String())
		if err != nil {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}
		mediaUploadRes = append(mediaUploadRes, MediaUploadRes{MediaID: m.String(), UploadURL: url})
	}

	if err := cp.sessionRepo.SetUploadPostSession(ctx, &createPostSession); err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}
	return mediaUploadRes, nil
}

func (cp *CreatePostUC) ExecuteCreatePostSession(ctx context.Context, cmd CreatePostSessionCmd) (CreatePostSessionRes, error) {
	createPostSession, err := domain.NewCreatePostSession(cmd.UserID, cmd.MediaCount)
	if err != nil {
		return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
	}

	var mediaUploadRes []MediaUploadRes
	for k, _ := range createPostSession.MediaIDs {
		url, err := cp.mediaStorage.GetPresignedURLUpload(ctx, k.String())
		if err != nil {
			return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
		}
		mediaUploadRes = append(mediaUploadRes, MediaUploadRes{MediaID: k.String(), UploadURL: url})

	}

	if err := cp.sessionRepo.SetUploadPostSession(ctx, createPostSession); err != nil {
		return CreatePostSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
	}

	return CreatePostSessionRes{
		SessionID:    createPostSession.SessionID.String(),
		MediaUploads: mediaUploadRes,
	}, nil
}

func (cp *CreatePostUC) ExecuteUploadPost(ctx context.Context, cmd UploadPostCmd) (*int64, error) {
	createPostSession, err := cp.sessionRepo.GetUploadPostSession(ctx, cmd.AuthorID, cmd.UpLoadSessionID)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	post, err := entity.NewPost(
		cmd.AuthorID,
		cmd.Visibility,
		cmd.Contents,
		cmd.AspectRatio,
	)
	if err != nil {
		return nil, err
	}

	cp.parseContent(ctx, post)

	aw, ah := post.AspectRatio().Demensions()
	mediaW := 1080
	mediaH := (ah * 1080) / aw

	thumbnailID := fmt.Sprintf("%s-thumbnail.jpeg", cmd.MediaUploads[0].MediaID.String())
	post.SetThumbnailID(thumbnailID)

	lenOfmedias := len(cmd.MediaUploads)
	var medias []entity.Media
	for i := 0; i < lenOfmedias; i++ {
		if _, ok := createPostSession.MediaIDs[cmd.MediaUploads[i].MediaID]; !ok {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}

		newMediaID := GenNewMediaID(cmd.MediaUploads[i].MediaID.String(), cmd.MediaUploads[i].Type)

		var taskType job.TaskType
		if cmd.MediaUploads[i].Type == entity.Image {
			taskType = job.ResizeImage
		} else {
			taskType = job.HLSVideo
		}

		mediaJob := job.ProcessMediaJob{
			ObjectSrc: job.ObjectLocation{
				Bucket:   config.TmpStorage,
				ObjectID: cmd.MediaUploads[i].MediaID.String(),
			},
			MediaType: job.MediaType(cmd.MediaUploads[i].Type),
			Tasks: []job.Task{
				job.Task{
					ObjectDes: job.ObjectLocation{
						Bucket:   config.PostsStorage,
						ObjectID: newMediaID,
					},
					Param: job.TaskParam{
						Width:  mediaW,
						Height: mediaH,
					},
					TaskType: taskType,
				},
			},
		}

		if i == 0 {
			mediaJob.Tasks = append(mediaJob.Tasks, job.Task{
				ObjectDes: job.ObjectLocation{
					Bucket:   config.PostsStorage,
					ObjectID: thumbnailID,
				},
				Param: job.TaskParam{
					Width:  300,
					Height: 400,
				},
				TaskType: job.ResizeImage,
			})
		}

		go func() {
			if err := cp.postProducer.SendProcessMediaJob(context.Background(), mediaJob); err != nil {
				log.Println(err)
			}
		}()

		medias = append(medias,
			entity.Media{
				MediaID:      newMediaID,
				Type:         cmd.MediaUploads[i].Type,
				DisplayOrder: i,
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

	go func() {
		if err := cp.postProducer.SendPostEvent(context.Background(), event.PostEvent{
			AuthorID: cmd.AuthorID,
			PostID:   *postId,
			Action:   event.Created,
		}); err != nil {
			log.Println(err)
		}
	}()

	return postId, nil
}

func GenNewMediaID(currentMediaID string, mediaType entity.MediaType) string {
	var extention string
	if mediaType == entity.Image {
		extention = ".jpeg"
	} else {
		extention = "/index.m3u8"
	}
	return fmt.Sprintf("%s%s", currentMediaID, extention)
}

func (cp *CreatePostUC) parseContent(ctx context.Context, p *entity.Post) {
	for i := range p.Contents() {
		switch p.Contents()[i].Type {
		case vo.Mention:
			_, username, _ := strings.Cut(p.Contents()[i].Text, "@")
			userID, err := cp.userRepo.GetIDByUsername(ctx, username)
			if err != nil {
				p.Contents()[i].Type = vo.Text
			} else {
				p.AppendMention(*userID)
				p.Contents()[i].Text = username
			}

		case vo.Hashtag:
			_, hashtag, _ := strings.Cut(p.Contents()[i].Text, "#")
			p.Appendhashtag(hashtag)
			p.Contents()[i].Text = hashtag
		}
	}
}
