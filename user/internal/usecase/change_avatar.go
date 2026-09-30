package usecase

import (
	"context"
	"fmt"
	"time"
	"user-service/config"
	"user-service/internal/domain"

	"github.com/google/uuid"
)

type ChangeAvatarUsecase struct {
	userRepo      domain.UserRepository
	cacheRepo     domain.CacheReposiotry
	objectStorage domain.ObjectStorage
	processImg    domain.ProcessImg
}

func NewChangeAvatarUsecase(
	userRepo domain.UserRepository,
	cacheRepo domain.CacheReposiotry,
	objectStorage domain.ObjectStorage,
	processImg domain.ProcessImg,
) *ChangeAvatarUsecase {
	return &ChangeAvatarUsecase{
		userRepo:      userRepo,
		cacheRepo:     cacheRepo,
		objectStorage: objectStorage,
		processImg:    processImg,
	}
}

func (cau *ChangeAvatarUsecase) ExecuteGetUrlUpload(ctx context.Context, userID int64) (string, error) {
	randID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteGetUrlUpload: %w", err)
	}

	presignedUrlUpload, err := cau.objectStorage.GetPresignedUrlUpload(ctx, string(config.TempBucket), randID.String(), time.Minute*10)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteGetUrlUpload: %w", err)
	}

	key := fmt.Sprintf("%s:%d", config.UserChangeAvatarPrefix, userID)
	if err := cau.cacheRepo.SetUploadAvatarSession(ctx, key, randID.String(), time.Minute*10); err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteGetUrlUpload: %w", err)
	}

	return presignedUrlUpload, nil
}

func (cau *ChangeAvatarUsecase) ExecuteCompletedUpload(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("%s:%d", config.UserChangeAvatarPrefix, userID)

	newAvatarID, err := cau.cacheRepo.GetUploadAvatarSession(ctx, key)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}

	avatarID, err := cau.userRepo.GetAvatarIDByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteGetUrlUpload: %w", err)
	}

	if avatarID == nil {
		id := fmt.Sprintf("%s.jpeg", newAvatarID)
		avatarID = &id
	}

	newAvatarReader, err := cau.objectStorage.GetObject(ctx, string(config.TempBucket), newAvatarID)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}

	thumbnailReader, size, err := cau.processImg.Resize(newAvatarReader, 600, 600)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}

	randID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}
	avatarThumbnailID := fmt.Sprintf("%s-thumbnail.jpeg", randID.String())

	if err := cau.objectStorage.Upload(ctx, string(config.AvatarBucket), avatarThumbnailID, thumbnailReader, "image/jpeg", int64(size)); err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}

	if err := cau.objectStorage.Copy(ctx, string(config.TempBucket), newAvatarID, string(config.AvatarBucket), *avatarID); err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}

	if err := cau.objectStorage.Remove(ctx, string(config.TempBucket), newAvatarID); err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}

	if err := cau.userRepo.UpdateAvatarIDAndAvatarThumbnailID(ctx, userID, *avatarID, avatarThumbnailID); err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}

	if err := cau.cacheRepo.RemoveSession(ctx, key); err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExecuteCompletedUpload: %w", err)
	}

	avatarURL := fmt.Sprintf("%s/%s/%s", config.StorageAddress, config.AvatarBucket, *avatarID)
	return avatarURL, nil
}
