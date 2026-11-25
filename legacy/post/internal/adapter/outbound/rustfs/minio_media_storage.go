package rustfsadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/htan06/Moments/post/internal/application/ports"
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

func (ms *MinIOMediaStorage) GetPresignedURLUpload(ctx context.Context, objKey []string) ([]ports.MediaUploadPresignedURL, error) {
	var mediaUploadsPresigned []ports.MediaUploadPresignedURL

	for _, k := range objKey {
		url, err := ms.presignedCli.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: &ms.tmpBucket,
			Key:    &k,
		}, s3.WithPresignExpires(30*time.Minute))

		if err != nil {
			return nil, fmt.Errorf("MinIOStorage.GetPresignedURLUpload: %w", err)
		}

		mediaUploadsPresigned = append(mediaUploadsPresigned, ports.MediaUploadPresignedURL{
			Key: k,
			URL: url.URL,
		})
	}

	return mediaUploadsPresigned, nil
}
