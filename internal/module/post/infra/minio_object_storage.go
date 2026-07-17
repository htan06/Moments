package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinIOStorage struct {
	conn *minio.Client
}

func NewMinIOStorage(conn *minio.Client) *MinIOStorage {
	return &MinIOStorage{
		conn: conn,
	}
}

func (ms *MinIOStorage) GetPresignedURLUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error) {
	url, err := ms.conn.PresignedPutObject(ctx, "tmp", objName, ttl)
	if err != nil {
		return "", fmt.Errorf("MinIOStorage.GetPresignedURLUpload: %w", err)
	}

	return url.String(), nil
}
