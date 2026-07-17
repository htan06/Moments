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

func (msr *MinIOStorage) GetPresignedUrlUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error) {
	url, err := msr.conn.PresignedPutObject(ctx, bucketName, objName, ttl)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (msr *MinIOStorage) PromoteAvatar(ctx context.Context, objName string) error {
	_, err := msr.conn.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: "avatars",
			Object: objName,
		},
		minio.CopySrcOptions{
			Bucket: "tmp",
			Object: objName,
		},
	)

	if err != nil {
		return fmt.Errorf("MinIOStorage.PromoteAvatar: %w", err)
	}

	if err := msr.conn.RemoveObject(ctx, "tmp", objName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("MinIOStorage.PromoteAvatar: %w", err)
	}
	return nil
}

func (msr *MinIOStorage) CheckObjectExists(ctx context.Context, bucketName string, objectName string) bool {
	_, err := msr.conn.StatObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return false
	}
	return true
}
