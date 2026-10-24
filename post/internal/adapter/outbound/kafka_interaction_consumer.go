package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/htan06/Moments/post/internal/application/domain/event"
	"github.com/segmentio/kafka-go"
)

type KafkaInteractionConsumer struct {
	reader      *kafka.Reader
	lastMessage kafka.Message
}

func NewKafkaInteractionConsumer(cfg *kafka.ReaderConfig) *KafkaInteractionConsumer {
	cfg.GroupID = "post-interaction"
	cfg.Topic = InteractionTopic

	return &KafkaInteractionConsumer{
		reader: kafka.NewReader(*cfg),
	}
}

func (k *KafkaInteractionConsumer) ReadMessage(ctx context.Context) (<-chan event.InteractionEvent, error) {
	eventChan := make(chan event.InteractionEvent, 10)

	go func(ctx context.Context, eventChan chan<- event.InteractionEvent, k *KafkaInteractionConsumer) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m, err := k.reader.FetchMessage(ctx)
				if err != nil {
					log.Println("KafkaInteractionConsumer.ReadMessage: %w", err)
					return
				}
				k.lastMessage = m

				var followEvent event.InteractionEvent
				if err := json.Unmarshal(m.Value, &followEvent); err != nil {
					log.Println("KafkaInteractionConsumer.ReadMessage: %w", err)
					return
				}
				eventChan <- followEvent
				log.Println(followEvent)
			}

		}

	}(ctx, eventChan, k)

	return eventChan, nil
}

func (k KafkaInteractionConsumer) Commit(ctx context.Context) error {
	if err := k.reader.CommitMessages(ctx, k.lastMessage); err != nil {
		return fmt.Errorf("KafkaInteractionConsumer.Commit: %w", err)
	}
	return nil
}
