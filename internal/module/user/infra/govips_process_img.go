package infra

import (
	"bytes"
	"fmt"
	"io"

	"github.com/davidbyttow/govips/v2/vips"
)

type ProcessImg struct {
}

func NewProcessImg() *ProcessImg {
	return &ProcessImg{}
}

func (p *ProcessImg) Resize(reader io.Reader, width int, height int) (io.Reader, int, error) {
	imgBuf, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, fmt.Errorf("ProcessImg.Resize: %w", err)
	}

	imgRef, err := vips.NewThumbnailFromBuffer(imgBuf, width, height, vips.InterestingAttention)
	if err != nil {
		return nil, 0, fmt.Errorf("ProcessImg.Resize: %w", err)
	}

	imgBytes, _, err := imgRef.ExportJpeg(&vips.JpegExportParams{Quality: 9})
	if err != nil {
		return nil, 0, fmt.Errorf("ProcessImg.Resize: %w", err)
	}

	return bytes.NewReader(imgBytes), len(imgBytes), nil
}
