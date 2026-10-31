package feature

import (
	"context"
	"fmt"
	"time"
	"user-service/config"
	"user-service/internal/application/domain/job"
	"user-service/internal/ports/outbound"

	"github.com/google/uuid"
)

type ChangeAvatar struct {
	userRepo      outbound.UserRepository
	cacheRepo     outbound.CacheReposiotry
	objectStorage outbound.ObjectStorage
	userProducer  outbound.UserProducer
}

func NewChangeAvatar(
	userRepo outbound.UserRepository,
	cacheRepo outbound.CacheReposiotry,
	objectStorage outbound.ObjectStorage,
	userProducer outbound.UserProducer,
) *ChangeAvatar {
	return &ChangeAvatar{
		userRepo:      userRepo,
		cacheRepo:     cacheRepo,
		objectStorage: objectStorage,
		userProducer:  userProducer,
	}
}

func (cau *ChangeAvatar) ExecuteGetUrlUpload(ctx context.Context, userID int64) (string, error) {
	randID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteGetUrlUpload: %w", err)
	}

	presignedUrlUpload, err := cau.objectStorage.GetPresignedUrlUpload(ctx, config.TmpStorage, randID.String(), time.Minute*5)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteGetUrlUpload: %w", err)
	}

	key := fmt.Sprintf("%s:%d", config.UserChangeAvatarPrefix, userID)
	if err := cau.cacheRepo.SetUploadAvatarSession(ctx, key, randID.String(), time.Minute*10); err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteGetUrlUpload: %w", err)
	}

	return presignedUrlUpload, nil
}

func (cau *ChangeAvatar) ExecuteCompletedUpload(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("%s:%d", config.UserChangeAvatarPrefix, userID)

	newAvatarID, err := cau.cacheRepo.GetUploadAvatarSession(ctx, key)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
	}

	oldAvatarID, err := cau.userRepo.GetAvatarKeyByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteGetUrlUpload: %w", err)
	}

	newAvatarKey := fmt.Sprintf("%s.jpeg", newAvatarID)
	newAvatarThumbnailKey := fmt.Sprintf("%s-thumbnail.jpeg", newAvatarID)

	avatarTask := job.Task{
		ObjectDes: job.ObjectLocation{
			Bucket:   config.AvatarsStorage,
			ObjectID: newAvatarKey,
		},
		TaskType: job.ResizeImage,
		Param: job.TaskParam{
			Width:  1000,
			Height: 1000,
		},
	}

	thumbnailTask :=
		job.Task{
			ObjectDes: job.ObjectLocation{
				Bucket:   config.AvatarsStorage,
				ObjectID: newAvatarThumbnailKey,
			},
			TaskType: job.ResizeImage,
			Param: job.TaskParam{
				Width:  600,
				Height: 600,
			},
		}

	job := job.ProcessMediaJob{
		ObjectSrc: job.ObjectLocation{
			Bucket:   config.TmpStorage,
			ObjectID: newAvatarID,
		},
		MediaType: job.Image,
		Tasks:     []job.Task{thumbnailTask, avatarTask},
	}

	if err := cau.userProducer.SendJob(ctx, job); err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
	}

	if oldAvatarID != nil {
		if err := cau.objectStorage.Remove(ctx, config.AvatarsStorage, *oldAvatarID); err != nil {
			return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
		}
	}

	if err := cau.userRepo.UpdateAvatarKeyAndAvatarThumbnailKey(ctx, userID, newAvatarKey, newAvatarThumbnailKey); err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
	}

	if err := cau.cacheRepo.RemoveSession(ctx, key); err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
	}

	avatarURL := fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, newAvatarKey)
	return avatarURL, nil
}
