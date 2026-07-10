package domain

import (
	"context"
)

type UserRepository interface {
	UpdateProfile(ctx context.Context, userID int64, fieldUpdates map[string]interface{}) (UserProfile, error)
	GetProfileByUsername(ctx context.Context, username string) (UserProfile, error)
}
