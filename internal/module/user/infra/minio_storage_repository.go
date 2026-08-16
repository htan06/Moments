package infra

import (
	"context"
	"fmt"
	"io"
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

func (msr *MinIOStorage) Copy(ctx context.Context, bucketSrc string, objSrc string, bucketDest string, objDest string) error {
	_, err := msr.conn.CopyObject(
		ctx,
		minio.CopyDestOptions{
			Bucket: bucketDest,
			Object: objDest,
		},
		minio.CopySrcOptions{
			Bucket: bucketSrc,
			Object: objSrc,
		},
	)

	if err != nil {
		return fmt.Errorf("MinIOStorage.PromoteAvatar: %w", err)
	}
	return nil
}
func (msr *MinIOStorage) GetObject(ctx context.Context, bucket string, key string) (io.Reader, error) {
	obj, err := msr.conn.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("MinIOStorage.GetObject: %w", err)
	}
	return obj, nil
}

func (msr *MinIOStorage) Upload(ctx context.Context, bucket string, key string, reader io.Reader, contentType string, size int64) error {
	if _, err := msr.conn.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return fmt.Errorf("MinIOStorage.Upload: %w", err)
	}
	return nil
}

func (msr *MinIOStorage) Remove(ctx context.Context, bucket string, key string) error {
	if err := msr.conn.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("MinIOStorage.Remove: %w", err)
	}
	return nil
}
