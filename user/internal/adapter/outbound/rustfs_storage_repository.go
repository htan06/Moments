package outbound

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type RustFSStorage struct {
	cli          *s3.Client
	presignedCli *s3.PresignClient
}

func NewRustFSStorage(cli *s3.Client) *RustFSStorage {
	return &RustFSStorage{
		cli:          cli,
		presignedCli: s3.NewPresignClient(cli),
	}
}

func (msr *RustFSStorage) GetPresignedUrlUpload(ctx context.Context, bucketName string, objName string, ttl time.Duration) (string, error) {
	url, err := msr.presignedCli.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: &bucketName, Key: &objName}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return url.URL, nil
}

func (msr *RustFSStorage) Remove(ctx context.Context, bucket string, key string) error {
	if _, err := msr.cli.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}); err != nil {
		return fmt.Errorf("RustFSStorage.Remove: %w", err)
	}
	return nil
}
