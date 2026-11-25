package feature

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/htan06/Moments/post/internal/application/ports"
	"github.com/htan06/Moments/post/internal/config"
	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/domain/entity"
	"github.com/htan06/Moments/post/internal/domain/job"
	"github.com/htan06/Moments/post/internal/domain/vo"
)

type MediaUpload struct {
	MediaKey  string `json:"media_key"`
	UploadURL string `json:"upload_url"`
}

type CreatePostUC struct {
	postCreator  ports.PostCreator
	userRepo     ports.UserRepository
	mediaStorage ports.MediaStorage
	sessionRepo  ports.SessionRepository
	postProducer ports.MediaJobPublisher
}

func NewCreatePostUC(
	postCreator ports.PostCreator,
	userRepo ports.UserRepository,
	mediaStorage ports.MediaStorage,
	sessionRepo ports.SessionRepository,
	postProducer ports.MediaJobPublisher,
) *CreatePostUC {
	return &CreatePostUC{
		postCreator:  postCreator,
		userRepo:     userRepo,
		mediaStorage: mediaStorage,
		sessionRepo:  sessionRepo,
		postProducer: postProducer,
	}
}

type CreateSessionCmd struct {
	UserID int64
}

type CreateSessionRes struct {
	SessionID uuid.UUID
}

func (cp *CreatePostUC) ExecuteCreateSession(ctx context.Context, cmd CreateSessionCmd) (CreateSessionRes, error) {
	createPostSession, err := entity.NewCreatePostSession(cmd.UserID)
	if err != nil {
		return CreateSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
	}

	if err := cp.sessionRepo.SetCreatePostSession(ctx, createPostSession); err != nil {
		return CreateSessionRes{}, fmt.Errorf("CreatePostUC.ExecutePrepareUploadPost: %w", err)
	}

	return CreateSessionRes{
		SessionID: createPostSession.SessionID(),
	}, nil
}

type RequestUploadURLsCmd struct {
	SessionID  uuid.UUID
	AuthorID   int64
	ImageCount int32
	VideoCount int32
}

type RequestUploadURLsResp struct {
	ImagesUploadPresigned []MediaUpload `json:"images"`
	VideosUploadPresigned []MediaUpload `json:"videos"`
}

func (cp *CreatePostUC) ExecuteRequestUploadURLs(ctx context.Context, cmd RequestUploadURLsCmd) (RequestUploadURLsResp, error) {
	createPostSession, err := cp.sessionRepo.GetCreatePostSession(ctx, cmd.AuthorID, cmd.SessionID.String())
	if err != nil {
		return RequestUploadURLsResp{}, fmt.Errorf("CreatePostUC.ExecuteRequestUploadURLs: %w", err)
	}

	uploadURLsResp := RequestUploadURLsResp{}
	if cmd.ImageCount > 0 {
		imagesKeys, err := createPostSession.RequestMediasKeys(cmd.ImageCount, entity.Image)
		if err != nil {
			return RequestUploadURLsResp{}, fmt.Errorf("CreatePostUC.ExecuteRequestUploadURLs: %w", err)
		}

		mediaUploadPresigneds, err := cp.mediaStorage.GetPresignedURLUpload(ctx, imagesKeys)
		if err != nil {
			return RequestUploadURLsResp{}, fmt.Errorf("CreatePostUC.ExecuteRequestUploadURLs: %w", err)
		}

		for _, mediaUpload := range mediaUploadPresigneds {
			uploadURLsResp.ImagesUploadPresigned = append(uploadURLsResp.ImagesUploadPresigned, MediaUpload{
				MediaKey:  mediaUpload.Key,
				UploadURL: mediaUpload.URL,
			})
		}
	}

	if cmd.VideoCount > 0 {
		videosKeys, err := createPostSession.RequestMediasKeys(cmd.ImageCount, entity.Video)
		if err != nil {
			return RequestUploadURLsResp{}, fmt.Errorf("CreatePostUC.ExecuteRequestUploadURLs: %w", err)
		}

		mediaUploadPresigneds, err := cp.mediaStorage.GetPresignedURLUpload(ctx, videosKeys)
		if err != nil {
			return RequestUploadURLsResp{}, fmt.Errorf("CreatePostUC.ExecuteRequestUploadURLs: %w", err)
		}

		for _, mediaUpload := range mediaUploadPresigneds {
			uploadURLsResp.VideosUploadPresigned = append(uploadURLsResp.VideosUploadPresigned, MediaUpload{
				MediaKey:  mediaUpload.Key,
				UploadURL: mediaUpload.URL,
			})
		}
	}
	if err := cp.sessionRepo.SetCreatePostSession(ctx, &createPostSession); err != nil {
		return RequestUploadURLsResp{}, fmt.Errorf("CreatePostUC.ExecuteRequestUploadURLs: %w", err)
	}
	return uploadURLsResp, nil
}

type UploadPostCmd struct {
	UpLoadSessionID string
	AuthorID        int64
	Content         vo.Content
	Visibility      entity.Visibility
	AspectRatio     vo.AspectRatio
	MediasKey       []string
}

func (cp *CreatePostUC) ExecuteUploadPost(ctx context.Context, cmd UploadPostCmd) (*int64, error) {
	createPostSession, err := cp.sessionRepo.GetCreatePostSession(ctx, cmd.AuthorID, cmd.UpLoadSessionID)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	thumbnailKey := fmt.Sprintf("%s-thumbnail.jpeg", cmd.MediasKey[0])

	var medias []entity.Media
	for i, m := range cmd.MediasKey {
		mediaType, exists := createPostSession.ContainsKey(m)
		if !exists {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", domain.ErrMediaTypeInvalid)
		}

		media, err := entity.NewMedia(m, mediaType, int32(i+1), 1, 1)
		if !exists {
			return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
		}

		medias = append(medias, *media)
	}

	post, err := entity.NewPost(
		createPostSession.UserID(),
		cmd.Visibility,
		cmd.Content,
		cmd.AspectRatio,
		thumbnailKey,
		medias,
		[]int64{},
		[]int64{},
	)

	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	postId, err := cp.postCreator.CreatePost(ctx, *post)
	if err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	if err := cp.publishMediaJob(post.ThumbnailKey(), post.Medias()); err != nil {
		return nil, fmt.Errorf("CreatePostUC.ExecuteCreatePost: %w", err)
	}

	return postId, nil
}

func genNewMediaKey(currentMediaID string, mediaType entity.MediaType) string {
	var extention string
	if mediaType == entity.Image {
		extention = ".jpeg"
	} else {
		extention = "/index.m3u8"
	}
	return fmt.Sprintf("%s%s", currentMediaID, extention)
}

func (cp *CreatePostUC) publishMediaJob(thumbnailKey string, medias []entity.Media) error {
	for i, m := range medias {

		var taskType job.TaskType
		if m.Type() == entity.Image {
			taskType = job.ResizeImage
		} else {
			taskType = job.HLSVideo
		}

		mediaJob := job.ProcessMediaJob{
			ObjectSrc: job.ObjectLocation{
				Bucket:   config.TmpStorage,
				ObjectID: m.MediaKey(),
			},
			MediaType: string(m.Type()),
			Tasks: []job.Task{
				job.Task{
					ObjectDes: job.ObjectLocation{
						Bucket:   config.PostsStorage,
						ObjectID: m.MediaKey(),
					},
					Param: job.TaskParam{
						Width:  1080,
						Height: 1080,
					},
					TaskType: taskType,
				},
			},
		}

		if i == 0 {
			mediaJob.Tasks = append(mediaJob.Tasks, job.Task{
				ObjectDes: job.ObjectLocation{
					Bucket:   config.PostsStorage,
					ObjectID: thumbnailKey,
				},
				Param: job.TaskParam{
					Width:  300,
					Height: 400,
				},
				TaskType: job.ResizeImage,
			})
		}

		if err := cp.postProducer.SendProcessMediaJob(context.Background(), mediaJob); err != nil {
			return fmt.Errorf("CreatePostUC.publishMediaJob: %w", err)
		}
	}
	return nil
}
