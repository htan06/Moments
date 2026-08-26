package infra

import (
	"bytes"
	"fmt"
	"io"

	"github.com/davidbyttow/govips/v2/vips"
)

type GoVipsProcessImg struct {
}

func NewGoVipsProcessImg() *GoVipsProcessImg {
	return &GoVipsProcessImg{}
}

func (p *GoVipsProcessImg) Resize(reader io.Reader, width int, height int) (io.Reader, int, error) {
	imgBuf, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("GoVipsProcessImg.Resize: %w", err)
	}

	imgRef, err := vips.NewThumbnailFromBuffer(imgBuf, width, height, vips.InterestingAttention)
	if err != nil {
		return nil, 0, fmt.Errorf("GoVipsProcessImg.Resize: %w", err)
	}

	imgBytes, _, err := imgRef.ExportJpeg(&vips.JpegExportParams{Quality: 90})
	if err != nil {
		return nil, 0, fmt.Errorf("GoVipsProcessImg.Resize: %w", err)
	}

	return bytes.NewReader(imgBytes), len(imgBytes), nil
}
