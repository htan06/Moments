package entity

import "time"

type MediaType string

const (
	Image MediaType = "IMAGE"
	Video MediaType = "VIDEO"
)

type Media struct {
	ID           int64
	MediaID      string
	Type         MediaType
	DisplayOrder int
	Width        int
	Height       int
	Duration     time.Duration
	Size         int
}
