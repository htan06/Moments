package outbound

import "context"

type UserRepository interface {
	GetIDByUsername(ctx context.Context, username string) (*int64, error)
}
