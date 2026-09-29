package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/htan06/Moments/post/config"
	"github.com/minio/minio-go/v7"
)

type MinIOMediaStorage struct {
	conn *minio.Client
}

func NewMinIOStorage(conn *minio.Client) *MinIOMediaStorage {
	return &MinIOMediaStorage{
		conn: conn,
	}
}

func (ms *MinIOMediaStorage) GetPresignedURLUpload(ctx context.Context, objName string) (string, error) {
	url, err := ms.conn.PresignedPutObject(ctx, string(config.TempBucket), objName, 1*time.Hour)
	if err != nil {
		return "", fmt.Errorf("MinIOStorage.GetPresignedURLUpload: %w", err)
	}

	return url.String(), nil
}
