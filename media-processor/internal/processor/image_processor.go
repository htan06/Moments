package processor

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/htan06/Moments/media-processor/internal/domain"
	"github.com/minio/minio-go/v7"
)

type ImageProcessor struct {
	minioConn *minio.Client
}

func NewImageProcessor(minioConn *minio.Client) *ImageProcessor {
	return &ImageProcessor{
		minioConn: minioConn,
	}
}

func (i *ImageProcessor) ProcessJob(ctx context.Context, job domain.ProcessMediaJob) error {
	srcReader, err := i.minioConn.GetObject(ctx, job.ObjectSrc.Bucket, job.ObjectSrc.ObjectID, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
	}

	check, err := check(srcReader, domain.Image)
	if err != nil {
		return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
	}
	if !check {
		return fmt.Errorf("ImageProcessor.ProcessJob: Media type is not image")
	}

	for _, task := range job.Tasks {
		width := task.Param.Width
		height := task.Param.Height

		if task.TaskType != domain.ResizeImage {
			return fmt.Errorf("ImageProcessor.ProcessJob: Cannot resolve task %s for %s", task.TaskType, job.MediaType)
		}

		newImgReader, size, err := i.resize(ctx, srcReader, width, height)
		if err != nil {
			return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
		}

		if _, err := i.minioConn.PutObject(
			ctx,
			task.ObjectDes.Bucket,
			task.ObjectDes.ObjectID,
			newImgReader,
			int64(size),
			minio.PutObjectOptions{ContentType: "image/jpeg"},
		); err != nil {
			return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
		}

		if _, err := srcReader.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
		}
	}
	return nil
}

func (i *ImageProcessor) resize(ctx context.Context, reader io.Reader, width int, height int) (io.Reader, int, error) {
	imgBuf, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("ImageProcessor.Resize: %w", err)
	}

	imgRef, err := vips.NewThumbnailFromBuffer(imgBuf, width, height, vips.InterestingAttention)
	if err != nil {
		return nil, 0, fmt.Errorf("ImageProcessor.Resize: %w", err)
	}

	imgBytes, _, err := imgRef.ExportJpeg(&vips.JpegExportParams{Quality: 90})
	if err != nil {
		return nil, 0, fmt.Errorf("ImageProcessor.Resize: %w", err)
	}

	return bytes.NewReader(imgBytes), len(imgBytes), nil
}
