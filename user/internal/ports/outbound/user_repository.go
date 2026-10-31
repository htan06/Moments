package outbound

import (
	"context"
	"io"
	"user-service/internal/application/domain/entity"
	"user-service/internal/application/domain/readmodel"
)

type UserRepository interface {
	CreateProfile(ctx context.Context, p entity.Profile) error
	GetAvatarKeyByUserID(ctx context.Context, userID int64) (*string, error)
	UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) error
	GetSelfProfileByUsername(ctx context.Context, username string) (readmodel.Profile, error)
	GetOtherProfileByUsername(ctx context.Context, currentUserID int64, targetUsername string) (readmodel.Profile, error)
	UpdateAvatarKeyAndAvatarThumbnailKey(ctx context.Context, userID int64, avatarID string, avatarThumbnailID string) error
	FindProfilesByUsername(ctx context.Context, username string) ([]readmodel.ProfileSummary, error)
	// IncPostCount(ctx context.Context, userID int64) error
	// DecPostCount(ctx context.Context, userID int64) error
	// IncFollowCount(ctx context.Context, followerID int64, followingID int64) error
	// DecFollowCount(ctx context.Context, followerID int64, followingID int64) error
}

type ProcessImg interface {
	Resize(reader io.Reader, width int, height int) (io.Reader, int, error)
}
