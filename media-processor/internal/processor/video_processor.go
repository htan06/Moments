package processor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/htan06/Moments/media-processor/internal/domain"
	"github.com/minio/minio-go/v7"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type VideoProcessor struct {
	minioConn *minio.Client
}

func NewVideoProcessor(minioConn *minio.Client) *VideoProcessor {
	return &VideoProcessor{
		minioConn: minioConn,
	}
}

func (v *VideoProcessor) ProcessJob(ctx context.Context, job domain.ProcessMediaJob) error {
	srcReader, err := v.minioConn.GetObject(ctx, job.ObjectSrc.Bucket, job.ObjectSrc.ObjectID, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
	}

	check, err := check(srcReader, domain.Video)
	if err != nil {
		return fmt.Errorf("VideoProcessor.ProcessJob: %w", err)
	}
	if !check {
		return fmt.Errorf("VideoProcessor.ProcessJob: Media type is not image")
	}

	for _, task := range job.Tasks {
		width := task.Param.Width
		height := task.Param.Height

		switch task.TaskType {
		case domain.HLSVideo:
			desFolder, err := v.hls(ctx, srcReader, job.ObjectSrc.ObjectID)
			if err != nil {
				return fmt.Errorf("VideoProcessor.ProcessJob: %w", err)
			}
			entries, err := os.ReadDir(desFolder)
			if err != nil {
				return fmt.Errorf("hlsVideo: %w", err)
			}
			var contentType string
			for _, entry := range entries {
				objName := fmt.Sprintf("%s/%s", job.ObjectSrc.ObjectID, entry.Name())
				objPath := fmt.Sprintf("%s/%s", desFolder, entry.Name())

				switch {
				case strings.HasSuffix(entry.Name(), ".m3u8"):
					contentType = "application/vnd.apple.mpegurl"
				case strings.HasSuffix(entry.Name(), ".ts"):
					contentType = "video/mp2t"
				default:
					contentType = "application/octet-stream"
				}

				if _, err := v.minioConn.FPutObject(ctx, task.ObjectDes.Bucket, objName, objPath, minio.PutObjectOptions{ContentType: contentType}); err != nil {
					return fmt.Errorf("hlsVideo: %w", err)
				}
			}

			if err := os.RemoveAll(desFolder); err != nil {
				return fmt.Errorf("hlsVideo: %w", err)
			}
		case domain.ResizeImage:
			newImgReader, size, err := v.thumbnail(ctx, srcReader, width, height)
			if err != nil {
				return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
			}

			if _, err := v.minioConn.PutObject(
				ctx,
				task.ObjectDes.Bucket,
				task.ObjectDes.ObjectID,
				newImgReader,
				int64(size),
				minio.PutObjectOptions{ContentType: "image/jpeg"},
			); err != nil {
				return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
			}

		}
		if _, err := srcReader.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("ImageProcessor.ProcessJob: %w", err)
		}
	}

	return nil
}

func (v *VideoProcessor) thumbnail(ctx context.Context, reader io.Reader, width int, height int) (io.Reader, int, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", "pipe:0",
		"-vf", "thumbnail",
		"-frames:v", "1",
		"-f", "mjpeg",
		"pipe:1")

	cmd.Stdin = reader

	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return nil, 0, fmt.Errorf("VideoProcessor.resize: %w", err)
	}

	imgRef, err := vips.NewThumbnailFromBuffer(buf.Bytes(), width, height, vips.InterestingAttention)
	if err != nil {
		return nil, 0, fmt.Errorf("VideoProcessor.resize: %w", err)
	}

	imgBytes, _, err := imgRef.ExportJpeg(&vips.JpegExportParams{Quality: 90})
	if err != nil {
		return nil, 0, fmt.Errorf("VideoProcessor.resize: %w", err)
	}

	return bytes.NewReader(imgBytes), len(imgBytes), nil
}

func (v *VideoProcessor) hls(ctx context.Context, reader io.Reader, objectID string) (string, error) {
	desFolder := fmt.Sprintf("././tmp/%s", objectID)

	if err := os.MkdirAll(desFolder, os.ModePerm); err != nil {
		return "", fmt.Errorf("hls: %w", err)
	}

	indexPath := fmt.Sprintf("%s/index.m3u8", desFolder)
	segmentPath := fmt.Sprintf("%s/segment_%%03d.ts", desFolder)

	if err := ffmpeg_go.Input("pipe:0", ffmpeg_go.KwArgs{}).
		Output(indexPath, ffmpeg_go.KwArgs{
			"c:v":                  "libx264",
			"c:a":                  "aac",
			"f":                    "hls",
			"hls_time":             4,
			"hls_playlist_type":    "vod",
			"hls_segment_filename": segmentPath,
		}).
		WithInput(reader).
		OverWriteOutput().
		Run(); err != nil {
		return "", fmt.Errorf("hls: %w", err)
	}
	return desFolder, nil
}
