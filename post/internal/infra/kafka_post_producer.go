package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/htan06/Moments/post/internal/domain"
	"github.com/segmentio/kafka-go"
)

const (
	InteractionTopic string = "interaction"
	PostTopic        string = "post"
)

type KafkaPostProducer struct {
	writer *kafka.Writer
}

func NewKafkaPostProducer() *KafkaPostProducer {
	return &KafkaPostProducer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      []string{"localhost:9092"},
			RequiredAcks: 1,
		}),
	}
}

func (k *KafkaPostProducer) SendPostEvent(ctx context.Context, postEvent domain.PostEvent) error {
	return k.send(ctx, PostTopic, postEvent)
}

func (k *KafkaPostProducer) SendInteractionEvent(ctx context.Context, interactionEvent domain.InteractionEvent) error {
	return k.send(ctx, InteractionTopic, interactionEvent)
}

func (k *KafkaPostProducer) send(ctx context.Context, topic string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("KafkaPostProducer.Send: %w", err)
	}

	err = k.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
		Topic: topic,
	})

	if err != nil {
		return fmt.Errorf("KafkaPostProducer.Send: %w", err)
	}
	return nil
}
