package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"user-service/internal/application/domain/job"

	"github.com/segmentio/kafka-go"
)

type KafakUserProducer struct {
	writer *kafka.Writer
}

func NewKafakUserProducer(cfg kafka.WriterConfig) *KafakUserProducer {
	cfg.Topic = "media-job"
	return &KafakUserProducer{
		writer: kafka.NewWriter(cfg),
	}
}

func (k *KafakUserProducer) SendJob(ctx context.Context, job job.ProcessMediaJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("KafakUserProducer.SendJob: %w", err)
	}

	if err := k.writer.WriteMessages(ctx, kafka.Message{
		Value: payload,
	}); err != nil {
		return fmt.Errorf("KafakUserProducer.SendJob: %w", err)
	}

	return nil
}
