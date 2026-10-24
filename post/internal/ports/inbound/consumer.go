package inbound

import (
	"context"

	"github.com/htan06/Moments/post/internal/application/domain/event"
)

type InteractionConsumer interface {
	ReadMessage(ctx context.Context) (<-chan event.InteractionEvent, error)
	Commit(ctx context.Context) error
}
