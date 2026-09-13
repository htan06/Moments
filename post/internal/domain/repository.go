package domain

import (
	"context"
	"io"
	"time"
)

type UserRepository interface {
	GetIDByUsername(ctx context.Context, username string) (*int64, error)
}

type PostRepository interface {
	CreatePost(ctx context.Context, post Post) (*int64, error)
	GetPost(ctx context.Context, postID int64) (PostReadModel, error)
	GetPostsByAuthorID(ctx context.Context, authorID int64, cursor int64, size int) ([]PostGridItem, error)
	DeletePostByUserIDAndPostID(ctx context.Context, userID int64, postID int64) error

	CreateLikePost(ctx context.Context, userID int64, postID int64) error
	DeleteLikePost(ctx context.Context, userID int64, postID int64) error

	GetRepostsByUsername(ctx context.Context, username string) ([]PostGridItem, error)
	CreateRepost(ctx context.Context, userID int64, postID int64) (int64, error)
	DeleteRepost(ctx context.Context, userID int64, repostID int64) error

	UpadateBatchLikeCount(ctx context.Context, list map[int64]int64) error
}

type ObjectStorage interface {
	GetPresignedURLDownload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error)
	GetPresignedURLUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error)
	GetObject(ctx context.Context, bucketName string, objName string) (io.ReadSeekCloser, error)
	PromotePostImage(ctx context.Context, objName string) error
	PutObject(ctx context.Context, bucketName string, objName string, file io.Reader, size int64) error
	FPutObject(ctx context.Context, bucketName string, objName string, filePath string, contentType string) error
}

type CacheRepository interface {
	SetUploadPostSession(ctx context.Context, userID int64, sessionID string, createPostSession *CreatePostSession) error
	GetUploadPostSession(ctx context.Context, userID int64, sessionID string) (CreatePostSession, error)
	IncPostLikes(ctx context.Context, userID int64, sessionID string) error
	DecPostLikes(ctx context.Context, userID int64, sessionID string) error
}

type ProcessImg interface {
	Resize(reader io.Reader, width int, height int) (io.Reader, int, error)
}

type ProcessVideo interface {
	HLS(url string, des string) error
}

type PostProducer interface {
	SendPostEvent(ctx context.Context, postEvent PostEvent) error
	SendInteractionEvent(ctx context.Context, interactionEvent InteractionEvent) error
}

type InteractionConsumer interface {
	ReadMessage(ctx context.Context) (<-chan InteractionEvent, error)
	Commit(ctx context.Context) error
}
