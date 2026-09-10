package infra

import (
	"fmt"
	"os"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type FfmpegProcessVideo struct {
}

func NewFfmpegProcessVideo() *FfmpegProcessVideo {
	return &FfmpegProcessVideo{}
}

func (p *FfmpegProcessVideo) HLS(url string, des string) error {
	desFolder := fmt.Sprintf("./internal/module/post/tmp/%s", des)

	if err := os.MkdirAll(desFolder, os.ModePerm); err != nil {
		return fmt.Errorf("FfmpegProcessVideo.HLS: %w", err)
	}

	indexPath := fmt.Sprintf("%s/index.m3u8", desFolder)
	segmentPath := fmt.Sprintf("%s/segment_%%03d.ts", desFolder)

	err := ffmpeg_go.Input(url, ffmpeg_go.KwArgs{}).
		Output(indexPath, ffmpeg_go.KwArgs{
			"c:v":                  "libx264",
			"c:a":                  "aac",
			"f":                    "hls",
			"hls_time":             4,
			"hls_playlist_type":    "vod",
			"hls_segment_filename": segmentPath,
		}).OverWriteOutput().Run()

	if err != nil {
		return fmt.Errorf("FfmpegProcessVideo.HLS: %w", err)
	}
	return nil
}
