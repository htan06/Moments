package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/htan06/Moments/internal/module/follow/domain"
	"github.com/segmentio/kafka-go"
)

type KafakFollowProducer struct {
	writer *kafka.Writer
}

func NewKafakFollowProducer() *KafakFollowProducer {
	return &KafakFollowProducer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      []string{"localhost:9092"},
			Topic:        "follow",
			RequiredAcks: 1,
		}),
	}
}

func (k *KafakFollowProducer) SendMessage(ctx context.Context, followEvent domain.FollowEvent) error {
	payload, err := json.Marshal(followEvent)
	if err != nil {
		return fmt.Errorf("KafakFollowProducer.SendMessage: %w", err)
	}

	if err := k.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	}); err != nil {
		return fmt.Errorf("KafakFollowProducer.SendMessage: %w", err)
	}
	
	return nil
}
