package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/segmentio/kafka-go"
)

type KafkaInteractionConsumer struct {
	reader      *kafka.Reader
	lastMessage kafka.Message
}

func NewKafkaInteractionConsumer() *KafkaInteractionConsumer {
	return &KafkaInteractionConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"localhost:9092"},
			GroupID: "post-interaction",
			Topic:   InteractionTopic,
		}),
	}
}

func (k *KafkaInteractionConsumer) ReadMessage(ctx context.Context) (<-chan domain.InteractionEvent, error) {
	eventChan := make(chan domain.InteractionEvent, 10)

	go func(ctx context.Context, eventChan chan<- domain.InteractionEvent, k *KafkaInteractionConsumer) {
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

				var followEvent domain.InteractionEvent
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
