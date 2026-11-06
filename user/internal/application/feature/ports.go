package feature

import (
	"context"
	"time"

	"github.com/htan06/Moments/user/internal/application/readmodel"
	"github.com/htan06/Moments/user/internal/domain"
)

type FollowCreator interface {
	CreateFollow(ctx context.Context, follow domain.Follow) error
}

type FollowReader interface {
	GetFollowing(ctx context.Context, username string, limit int32, cursor int64) ([]readmodel.UserSummary, error)
	GetFollowers(ctx context.Context, username string, limit int32, cursor int64) ([]readmodel.UserSummary, error)
}

type FollowDeleter interface {
	RemoveByIDAndFollowerID(ctx context.Context, f domain.Follow) error
	RemoveByIDAndFollowingID(ctx context.Context, f domain.Follow) error
}

type FollowProducer interface {
	PublishFollowCreated(ctx context.Context, e domain.FollowCreated) error
	PublishFollowDeleted(ctx context.Context, e domain.FollowDeleted) error
}

type CacheReposiotry interface {
	SetUploadAvatarSession(ctx context.Context, key string, avatarName string, ttl time.Duration) error
	GetUploadAvatarSession(ctx context.Context, key string) (string, error)
	RemoveSession(ctx context.Context, key string) error
}

type ObjectStorage interface {
	GetPresignedUrlUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error)
	Remove(ctx context.Context, bucket string, key string) error
}

type ProfileCreator interface {
	CreateProfile(ctx context.Context, p domain.Profile) error
}

type ProfileUpdater interface {
	UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) error
	UpdateAvatarKeyAndAvatarThumbnailKey(ctx context.Context, userID int64, avatarID string, avatarThumbnailID string) error
}

type ProfileReader interface {
	GetAvatarKeyByUserID(ctx context.Context, userID int64) (*string, error)
	GetSelfProfileByUsername(ctx context.Context, username string) (readmodel.Profile, error)
	GetOtherProfileByUsername(ctx context.Context, currentUserID int64, targetUsername string) (readmodel.Profile, error)
	FindProfilesByUsername(ctx context.Context, username string) ([]readmodel.ProfileSummary, error)
	GetProfilesSummary(ctx context.Context, currentUserId int64, ids []int64) ([]readmodel.ProfileSummary, error)
}

type UserProducer interface {
	SendJob(ctx context.Context, job domain.ProcessMediaJob) error
}
