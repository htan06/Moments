package domain

import "context"

type MediaRepository interface {
	Update(ctx context.Context, id int64, mediaID string) error
}
