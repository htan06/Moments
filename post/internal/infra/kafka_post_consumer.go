package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/segmentio/kafka-go"
)

type KafkaInteractionConsumer struct {
	reader *kafka.Reader
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

func (k *KafkaInteractionConsumer) ReadMessage(ctx context.Context) (domain.InteractionEvent, error) {
	m, err := k.reader.ReadMessage(ctx)
	if err != nil {
		return domain.InteractionEvent{}, fmt.Errorf("KafkaInteractionConsumer.ReadMessage: %w", err)
	}
	defer k.reader.CommitMessages(ctx, m)

	var followEvent domain.InteractionEvent
	if err := json.Unmarshal(m.Value, &followEvent); err != nil {
		return domain.InteractionEvent{}, fmt.Errorf("KafkaInteractionConsumer.ReadMessage: %w", err)
	}

	return followEvent, nil
}
