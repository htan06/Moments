package domain

import (
	"context"
	"io"
)

type UserRepository interface {
	GetIDByUsername(ctx context.Context, username string) (*int64, error)
}

type PostRepository interface {
	CreatePost(ctx context.Context, post Post) (*int64, error)
	GetPost(ctx context.Context, postID int64) (PostReadModel, error)
	GetPostsByAuthorID(ctx context.Context, viewerID int64, authorID int64, cursor int64, size int) ([]PostSummary, error)
	DeletePostByUserIDAndPostID(ctx context.Context, userID int64, postID int64) error

	CreateLikePost(ctx context.Context, userID int64, postID int64) error
	DeleteLikePost(ctx context.Context, userID int64, postID int64) error

	GetRepostsByUsername(ctx context.Context, username string) ([]PostSummary, error)
	CreateRepost(ctx context.Context, userID int64, postID int64) (int64, error)
	DeleteRepost(ctx context.Context, userID int64, repostID int64) error

	UpadateBatchLikeCount(ctx context.Context, list map[int64]int64) error
	GetLikeCount(ctx context.Context, postID int64) (int64, error)
}

type MediaStorage interface {
	GetPresignedURLUpload(ctx context.Context, objName string) (string, error)
}

type CacheRepository interface {
	SetUploadPostSession(ctx context.Context, userID int64, sessionID string, createPostSession *CreatePostSession) error
	GetUploadPostSession(ctx context.Context, userID int64, sessionID string) (CreatePostSession, error)
	IncPostLikes(ctx context.Context, postID int64) error
	SetPostLikesIfNotExists(ctx context.Context, postID int64, likeCount int) error
	DecPostLikes(ctx context.Context, postID int64) error
	GetLikeCount(ctx context.Context, postIDs ...int64) (map[int64]int64, error)
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
	SendProcessMediaJob(ctx context.Context, job interface{}) error
}

type InteractionConsumer interface {
	ReadMessage(ctx context.Context) (<-chan InteractionEvent, error)
	Commit(ctx context.Context) error
}
