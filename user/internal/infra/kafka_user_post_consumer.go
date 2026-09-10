package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"user-service/internal/domain"

	kafka "github.com/segmentio/kafka-go"
)

type KafkaUserPostConsumer struct {
	reader *kafka.Reader
}

func NewKafkaUserPostConsumer() *KafkaUserPostConsumer {
	return &KafkaUserPostConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "post",
			GroupID: "user-consumer",
		}),
	}
}

func (k *KafkaUserPostConsumer) ReadMessage(ctx context.Context) (domain.PostEvent, error) {
	m, err := k.reader.ReadMessage(ctx)
	if err != nil {
		return domain.PostEvent{}, fmt.Errorf("KafkaUserPostConsumer.Read: %w", err)
	}
	defer k.reader.CommitMessages(ctx, m)

	var PostEvent domain.PostEvent
	if err := json.Unmarshal(m.Value, &PostEvent); err != nil {
		return domain.PostEvent{}, fmt.Errorf("KafkaUserPostConsumer.Read: %w", err)
	}

	return PostEvent, nil
}
