package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/domain"
)

type ChangeAvatarUsecase struct {
	userRepo      domain.UserRepository
	cacheRepo     domain.CacheReposiotry
	objectStorage domain.ObjectStorage
}

func NewChangeAvatarUsecase(
	userRepo domain.UserRepository,
	cacheRepo domain.CacheReposiotry,
	objectStorage domain.ObjectStorage,
) *ChangeAvatarUsecase {
	return &ChangeAvatarUsecase{
		userRepo:      userRepo,
		cacheRepo:     cacheRepo,
		objectStorage: objectStorage,
	}
}

func (cau *ChangeAvatarUsecase) ExcuteGetUrlUpload(ctx context.Context, userID int64) (string, error) {
	avatarID, err := cau.userRepo.GetAvatarIDByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExcuteGetUrlUpload: %w", err)
	}

	if avatarID == "" {
		randID, err := uuid.NewRandom()
		if err != nil {
			return "", fmt.Errorf("ChangeAvatarUsecase.ExcuteGetUrlUpload: %w", err)
		}
		avatarID = randID.String()
	}

	presignedUrlUpload, err := cau.objectStorage.GetPresignedUrlUpload(ctx, "tmp", avatarID, time.Minute*10)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExcuteGetUrlUpload: %w", err)
	}

	key := fmt.Sprintf("upload-session:%d", userID)
	if err := cau.cacheRepo.SetUploadAvatarSession(ctx, key, avatarID, time.Minute*10); err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExcuteGetUrlUpload: %w", err)
	}

	return presignedUrlUpload, nil
}

func (cau *ChangeAvatarUsecase) ExcuteCompletedUpload(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("upload-session:%d", userID)

	avatarID, err := cau.cacheRepo.GetUploadAvatarSession(ctx, key)
	if err != nil {
		return fmt.Errorf("ChangeAvatarUsecase.ExcuteCompletedUpload: %w", err)
	}

	if err := cau.objectStorage.PromoteAvatar(ctx, avatarID); err != nil {
		return fmt.Errorf("ChangeAvatarUsecase.ExcuteCompletedUpload: %w", err)
	}

	if err := cau.userRepo.UpdateAvatarID(ctx, userID, avatarID); err != nil {
		return fmt.Errorf("ChangeAvatarUsecase.ExcuteCompletedUpload: %w", err)
	}

	if err := cau.cacheRepo.RemoveSession(ctx, key); err != nil {
		return fmt.Errorf("ChangeAvatarUsecase.ExcuteCompletedUpload: %w", err)
	}
	return nil
}
