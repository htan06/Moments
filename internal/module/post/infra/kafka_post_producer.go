package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/htan06/Moments/internal/module/post/domain"
	"github.com/segmentio/kafka-go"
)

type KafkaPostProducer struct {
	writer *kafka.Writer
}

func NewKafkaPostProducer() *KafkaPostProducer {
	return &KafkaPostProducer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      []string{"localhost:9092"},
			Topic:        "post",
			RequiredAcks: 1,
		}),
	}
}

func (k *KafkaPostProducer) Send(ctx context.Context, postCreated domain.PostEvent) error {
	payload, err := json.Marshal(postCreated)
	if err != nil {
		return fmt.Errorf("KafkaPostProducer.Send: %w", err)
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	})

	if err != nil {
		return fmt.Errorf("KafkaPostProducer.Send: %w", err)
	}
	return nil
}
