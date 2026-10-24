package readmodel

import "time"

type Media struct {
	ID           int64     `json:"id" db:"m_id"`
	Type         string    `json:"type" db:"type"`
	MediaID      string    `json:"media_url" db:"media_id"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	Width        *int      `json:"width,omitempty" db:"width"`
	Height       *int      `json:"height,omitempty" db:"height"`
	Duration     *int      `json:"duration,omitempty" db:"duration"`
	Size         *int64    `json:"size,omitempty" db:"size"`
	CreatedAt    time.Time `json:"created_at" db:"m_created_at"`
}
