package outbound

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MinIOMediaStorage struct {
	cli          *s3.Client
	presignedCli *s3.PresignClient
	tmpBucket    string
	postsBucket  string
}

func NewMinIOStorage(cli *s3.Client) *MinIOMediaStorage {
	return &MinIOMediaStorage{
		cli:          cli,
		presignedCli: s3.NewPresignClient(cli),
		tmpBucket:    "tmp",
		postsBucket:  "posts",
	}
}

func (ms *MinIOMediaStorage) GetPresignedURLUpload(ctx context.Context, objName string) (string, error) {
	url, err := ms.presignedCli.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &ms.tmpBucket,
		Key:    &objName,
	}, s3.WithPresignExpires(30*time.Minute))
	if err != nil {
		return "", fmt.Errorf("MinIOStorage.GetPresignedURLUpload: %w", err)
	}

	return url.URL, nil
}
