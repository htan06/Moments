package outbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/htan06/Moments/user/internal/domain"
	"github.com/segmentio/kafka-go"
)

type KafakUserProducer struct {
	writer *kafka.Writer
}

func NewKafakUserProducer(cfg kafka.WriterConfig) *KafakUserProducer {
	cfg.Topic = "media-domain"
	return &KafakUserProducer{
		writer: kafka.NewWriter(cfg),
	}
}

func (k *KafakUserProducer) SendJob(ctx context.Context, domain domain.ProcessMediaJob) error {
	payload, err := json.Marshal(domain)
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
