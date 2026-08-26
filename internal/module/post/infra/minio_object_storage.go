package infra

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/htan06/Moments/internal/config"
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

func (ms *MinIOStorage) GetPresignedURLDownload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error) {
	url, err := ms.conn.PresignedGetObject(ctx, bucketName, objName, ttl, url.Values{})
	if err != nil {
		return "", fmt.Errorf("MinIOStorage.GetPresignedURLUpload: %w", err)
	}

	return url.String(), nil
}

func (ms *MinIOStorage) GetPresignedURLUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error) {
	url, err := ms.conn.PresignedPutObject(ctx, bucketName, objName, ttl)
	if err != nil {
		return "", fmt.Errorf("MinIOStorage.GetPresignedURLUpload: %w", err)
	}

	return url.String(), nil
}

func (ms *MinIOStorage) GetObject(ctx context.Context, bucketName string, objName string) (io.ReadSeekCloser, error) {
	obj, err := ms.conn.GetObject(ctx, bucketName, objName, minio.GetObjectOptions{})

	if err != nil {
		return nil, fmt.Errorf("MinIOStorage.GetObject: %w", err)
	}

	return obj, nil
}

func (ms *MinIOStorage) PromotePostImage(ctx context.Context, objName string) error {
	_, err := ms.conn.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: string(config.PostBucket),
			Object: objName,
		},
		minio.CopySrcOptions{
			Bucket: string(config.TempBucket),
			Object: objName,
		})

	if err != nil {
		return fmt.Errorf("MinIOStorage.PromotePostImage: %w", err)
	}

	if err := ms.conn.RemoveObject(ctx, string(config.TempBucket), objName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("MinIOStorage.PromotePostImage: %w", err)
	}
	return nil
}

func (ms *MinIOStorage) PutObject(ctx context.Context, bucketName string, objName string, file io.Reader, size int64) error {
	_, err := ms.conn.PutObject(ctx, bucketName, objName, file, size, minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		return fmt.Errorf("MinIOStorage.PutObject: %w", err)
	}
	return nil
}

func (ms *MinIOStorage) FPutObject(ctx context.Context, bucketName string, objName string, filePath string, contentType string) error {
	_, err := ms.conn.FPutObject(ctx, bucketName, objName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("MinIOStorage.PutObject: %w", err)
	}
	return nil
}