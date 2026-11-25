package outbound

import (
	"context"

	"github.com/htan06/Moments/post/internal/application/domain/event"
)

type PostProducer interface {
	SendPostEvent(ctx context.Context, postEvent event.PostEvent) error
	SendInteractionEvent(ctx context.Context, interactionEvent event.InteractionEvent) error
	SendProcessMediaJob(ctx context.Context, job interface{}) error
}
