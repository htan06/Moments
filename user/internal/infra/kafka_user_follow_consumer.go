package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"user-service/internal/domain"

	kafka "github.com/segmentio/kafka-go"
)

type KafkaUserFollowConsumer struct {
	reader *kafka.Reader
}

func NewKafkaUserFollowConsumer() *KafkaUserFollowConsumer {
	return &KafkaUserFollowConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "follow",
			GroupID: "user-follow",
		}),
	}
}

func (k *KafkaUserFollowConsumer) ReadMessage(ctx context.Context) (domain.FollowEvent, error) {
	m, err := k.reader.ReadMessage(ctx)
	if err != nil {
		return domain.FollowEvent{}, fmt.Errorf("KafkaUserFollowConsumer.ReadMessage: %w", err)
	}

	var followEvent domain.FollowEvent
	if err := json.Unmarshal(m.Value, &followEvent); err != nil {
		return domain.FollowEvent{}, fmt.Errorf("KafkaUserFollowConsumer.ReadMessage: %w", err)
	}

	return followEvent, nil
}
