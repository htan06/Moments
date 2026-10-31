package outbound

import (
	"context"
	"time"
)

type ObjectStorage interface {
	GetPresignedUrlUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error)
	Remove(ctx context.Context, bucket string, key string) error
}
