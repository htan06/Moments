package outbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/htan06/Moments/post/internal/application/domain/event"
	"github.com/segmentio/kafka-go"
)

const (
	InteractionTopic string = "interaction"
	PostTopic        string = "post"
	MediaJobTopic    string = "media-job"
)

type KafkaPostProducer struct {
	writer *kafka.Writer
}

func NewKafkaPostProducer(cfg *kafka.WriterConfig) *KafkaPostProducer {
	cfg.RequiredAcks = 1

	return &KafkaPostProducer{
		writer: kafka.NewWriter(*cfg),
	}
}

func (k *KafkaPostProducer) SendPostEvent(ctx context.Context, postEvent event.PostEvent) error {
	return k.send(ctx, PostTopic, postEvent)
}

func (k *KafkaPostProducer) SendInteractionEvent(ctx context.Context, interactionEvent event.InteractionEvent) error {
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

func (k *KafkaPostProducer) SendProcessMediaJob(ctx context.Context, job interface{}) error {
	return k.send(ctx, MediaJobTopic, job)
}
