package outbound

import (
	"context"
	"fmt"
	"user-service/internal/application/domain/event"

	kafka "github.com/segmentio/kafka-go"
)

type KafakFollowProducer struct {
	writer             *kafka.Writer
	followCreatedTopic string
	followDeletedTopic string
}

func NewKafakFollowProducer(cfg kafka.WriterConfig) *KafakFollowProducer {
	cfg.RequiredAcks = 1
	return &KafakFollowProducer{
		writer:             kafka.NewWriter(cfg),
		followCreatedTopic: "follow-created",
		followDeletedTopic: "follow-deleted",
	}
}

func (k *KafakFollowProducer) sendMessage(ctx context.Context, topic string, data []byte) error {
	if err := k.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
		Topic: topic,
	}); err != nil {
		return fmt.Errorf("KafakFollowProducer.SendMessage: %w", err)
	}

	return nil
}

func (k *KafakFollowProducer) PublishFollowCreated(ctx context.Context, e event.FollowCreated) error {
	payload, err := e.MarshalJSON()
	if err != nil {
		return fmt.Errorf("KafakFollowProducer.PublishFollowCreated: %w", err)
	}
	return k.sendMessage(ctx, k.followCreatedTopic, payload)
}

func (k *KafakFollowProducer) PublishFollowDeleted(ctx context.Context, e event.FollowDeleted) error {
	payload, err := e.MarshalJSON()
	if err != nil {
		return fmt.Errorf("KafakFollowProducer.PublishFollowDeleted: %w", err)
	}
	return k.sendMessage(ctx, k.followDeletedTopic, payload)
}
