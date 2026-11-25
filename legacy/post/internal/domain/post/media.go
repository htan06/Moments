package post

import (
	"strings"
	"time"

	"github.com/htan06/Moments/post/internal/domain"
)

// Media is a child entity of Post, identified by its mediaKey.
type Media struct {
	mediaKey     string
	mediaType    MediaType
	displayOrder int32
	status       MediaStatus
	width        int32
	height       int32
	createdAt    time.Time
	updatedAt    time.Time
}

// NewMedia creates a just-uploaded media in PROCESSING; its dimensions
// are only known once it is processed (MarkMediaPublished).
func NewMedia(mediaKey string, mediaType MediaType, displayOrder int32) (Media, error) {
	now := time.Now().UTC()
	m := Media{
		mediaKey:     mediaKey,
		mediaType:    mediaType,
		displayOrder: displayOrder,
		status:       MediaProcessing,
		createdAt:    now,
		updatedAt:    now,
	}
	if err := m.validate(); err != nil {
		return Media{}, err
	}
	return m, nil
}

type RehydrateMediaParams struct {
	MediaKey     string
	MediaType    MediaType
	DisplayOrder int32
	Status       MediaStatus
	Width        int32
	Height       int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RehydrateMedia rebuilds a Media from persistence.
func RehydrateMedia(p RehydrateMediaParams) (Media, error) {
	m := Media{
		mediaKey:     p.MediaKey,
		mediaType:    p.MediaType,
		displayOrder: p.DisplayOrder,
		status:       p.Status,
		width:        p.Width,
		height:       p.Height,
		createdAt:    p.CreatedAt,
		updatedAt:    p.UpdatedAt,
	}
	if err := m.validate(); err != nil {
		return Media{}, err
	}
	return m, nil
}

func (m Media) MediaKey() string     { return m.mediaKey }
func (m Media) MediaType() MediaType { return m.mediaType }
func (m Media) DisplayOrder() int32  { return m.displayOrder }
func (m Media) Status() MediaStatus  { return m.status }
func (m Media) Width() int32         { return m.width }
func (m Media) Height() int32        { return m.height }
func (m Media) CreatedAt() time.Time { return m.createdAt }
func (m Media) UpdatedAt() time.Time { return m.updatedAt }
func (m Media) IsPublished() bool    { return m.status == MediaPublished }

func (m *Media) markPublished(width, height int32, now time.Time) error {
	if m.status == MediaPublished {
		return domain.ErrMediaAlreadyPublished
	}
	if width <= 0 {
		return domain.ErrWidthInvalid
	}
	if height <= 0 {
		return domain.ErrHeightInvalid
	}
	m.width, m.height = width, height
	m.status = MediaPublished
	m.updatedAt = now
	return nil
}

func (m Media) validate() error {
	if strings.TrimSpace(m.mediaKey) == "" {
		return domain.ErrMediaKeyEmpty
	}
	if !m.mediaType.Valid() {
		return domain.ErrMediaTypeInvalid
	}
	if m.displayOrder <= 0 {
		return domain.ErrDisplayOrderInvalid
	}
	if !m.status.Valid() {
		return domain.ErrMediaStatusInvalid
	}
	if m.width < 0 {
		return domain.ErrWidthInvalid
	}
	if m.height < 0 {
		return domain.ErrHeightInvalid
	}
	// INV-MEDIA-01
	if m.status == MediaPublished {
		if m.width == 0 {
			return domain.ErrWidthInvalid
		}
		if m.height == 0 {
			return domain.ErrHeightInvalid
		}
	}
	return nil
}
