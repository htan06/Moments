package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/htan06/echo-messenger-rest-api/internal/config"
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
	var avatarID string
	avatarIDptr, err := cau.userRepo.GetAvatarIDByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExcuteGetUrlUpload: %w", err)
	}

	if avatarIDptr == nil {
		randID, err := uuid.NewRandom()
		if err != nil {
			return "", fmt.Errorf("ChangeAvatarUsecase.ExcuteGetUrlUpload: %w", err)
		}
		avatarID = randID.String()
	} else {
		avatarID = *avatarIDptr
	}

	presignedUrlUpload, err := cau.objectStorage.GetPresignedUrlUpload(ctx, string(config.TempBucket), avatarID, time.Minute*10)
	if err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExcuteGetUrlUpload: %w", err)
	}

	key := fmt.Sprintf("%s:%d", config.UserChangeAvatarPrefix, userID)
	if err := cau.cacheRepo.SetUploadAvatarSession(ctx, key, avatarID, time.Minute*10); err != nil {
		return "", fmt.Errorf("ChangeAvatarUsecase.ExcuteGetUrlUpload: %w", err)
	}

	return presignedUrlUpload, nil
}

func (cau *ChangeAvatarUsecase) ExcuteCompletedUpload(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("%s:%d", config.UserChangeAvatarPrefix, userID)

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
