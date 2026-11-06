package feature

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/htan06/Moments/user/config"
	"github.com/htan06/Moments/user/internal/domain"
)

type ChangeAvatar struct {
	profileUpdater ProfileUpdater
	profileFinder  ProfileReader
	cacheRepo      CacheReposiotry
	objectStorage  ObjectStorage
	userProducer   UserProducer
}

func NewChangeAvatar(
	profileUpdater ProfileUpdater,
	profileFinder ProfileReader,
	cacheRepo CacheReposiotry,
	objectStorage ObjectStorage,
	userProducer UserProducer,
) *ChangeAvatar {
	return &ChangeAvatar{
		profileUpdater: profileUpdater,
		profileFinder:  profileFinder,
		cacheRepo:      cacheRepo,
		objectStorage:  objectStorage,
		userProducer:   userProducer,
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

	oldAvatarID, err := cau.profileFinder.GetAvatarKeyByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteGetUrlUpload: %w", err)
	}

	newAvatarKey := fmt.Sprintf("%s.jpeg", newAvatarID)
	newAvatarThumbnailKey := fmt.Sprintf("%s-thumbnail.jpeg", newAvatarID)

	avatarTask := domain.Task{
		ObjectDes: domain.ObjectLocation{
			Bucket:   config.AvatarsStorage,
			ObjectID: newAvatarKey,
		},
		TaskType: domain.ResizeImage,
		Param: domain.TaskParam{
			Width:  1000,
			Height: 1000,
		},
	}

	thumbnailTask :=
		domain.Task{
			ObjectDes: domain.ObjectLocation{
				Bucket:   config.AvatarsStorage,
				ObjectID: newAvatarThumbnailKey,
			},
			TaskType: domain.ResizeImage,
			Param: domain.TaskParam{
				Width:  600,
				Height: 600,
			},
		}

	job := domain.ProcessMediaJob{
		ObjectSrc: domain.ObjectLocation{
			Bucket:   config.TmpStorage,
			ObjectID: newAvatarID,
		},
		MediaType: domain.Image,
		Tasks:     []domain.Task{thumbnailTask, avatarTask},
	}

	if err := cau.userProducer.SendJob(ctx, job); err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
	}

	if oldAvatarID != nil {
		if err := cau.objectStorage.Remove(ctx, config.AvatarsStorage, *oldAvatarID); err != nil {
			return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
		}
	}

	if err := cau.profileUpdater.UpdateAvatarKeyAndAvatarThumbnailKey(ctx, userID, newAvatarKey, newAvatarThumbnailKey); err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
	}

	if err := cau.cacheRepo.RemoveSession(ctx, key); err != nil {
		return "", fmt.Errorf("ChangeAvatar.ExecuteCompletedUpload: %w", err)
	}

	avatarURL := fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarsStorage, newAvatarKey)
	return avatarURL, nil
}
