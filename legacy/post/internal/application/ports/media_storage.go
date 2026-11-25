package ports

import "context"

type MediaUploadPresignedURL struct {
	Key string
	URL string
}

type MediaStorage interface {
	GetPresignedURLUpload(ctx context.Context, objKey []string) ([]MediaUploadPresignedURL, error)
}
