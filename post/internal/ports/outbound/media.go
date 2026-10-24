package outbound

import "context"

type MediaStorage interface {
	GetPresignedURLUpload(ctx context.Context, objName string) (string, error)
}
