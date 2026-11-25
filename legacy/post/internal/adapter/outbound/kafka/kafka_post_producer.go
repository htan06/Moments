package kafkaadpter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/htan06/Moments/post/internal/domain/event"
	"github.com/htan06/Moments/post/internal/domain/job"
	"github.com/segmentio/kafka-go"
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

func (k *KafkaPostProducer) SendPostCreated(ctx context.Context, postEvent event.PostCreated) error {
	return k.send(ctx, PostCreatedTopic, postEvent)
}

func (k *KafkaPostProducer) SendPostDeleted(ctx context.Context, postEvent event.PostDeleted) error {
	return k.send(ctx, PostDeletedTopic, postEvent)
}

func (k *KafkaPostProducer) SendPostLiked(ctx context.Context, postEvent event.PostLiked) error {
	return k.send(ctx, PostLikedTopic, postEvent)
}

func (k *KafkaPostProducer) SendPostUnliked(ctx context.Context, postEvent event.PostUnliked) error {
	return k.send(ctx, PostUnlikedTopic, postEvent)
}

func (k *KafkaPostProducer) send(ctx context.Context, topic string, event any) error {
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

func (k *KafkaPostProducer) SendProcessMediaJob(ctx context.Context, job job.ProcessMediaJob) error {
	return k.send(ctx, MediaJobTopic, job)
}
