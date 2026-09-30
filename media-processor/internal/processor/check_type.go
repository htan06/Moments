package processor

import (
	"fmt"
	"io"

	"github.com/h2non/filetype"
	"github.com/htan06/Moments/media-processor/internal/domain"
)

func check(reader io.ReadSeeker, mediaType domain.MediaType) (bool, error) {

	head := make([]byte, 262)
	if _, err := reader.Read(head); err != nil {
		return false, fmt.Errorf("Check: %w", err)
	}

	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return false, fmt.Errorf("Check: %w", err)
	}

	if mediaType == domain.Image {
		return filetype.IsImage(head), nil
	} else if mediaType == domain.Video {
		return filetype.IsVideo(head), nil
	} else {
		return false, fmt.Errorf("File not a image or video")
	}
}
