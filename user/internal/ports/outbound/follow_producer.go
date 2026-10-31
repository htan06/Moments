package outbound

import (
	"context"
	"user-service/internal/application/domain/event"
)

type FollowProducer interface {
	PublishFollowCreated(ctx context.Context, e event.FollowCreated) error
	PublishFollowDeleted(ctx context.Context, e event.FollowDeleted) error
}
