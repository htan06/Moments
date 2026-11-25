package inbound

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"user-service/internal/application/domain/event"

// 	kafka "github.com/segmentio/kafka-go"
// )

// type KafkaUserFollowConsumer struct {
// 	reader *kafka.Reader
// }

// func NewKafkaUserFollowConsumer() *KafkaUserFollowConsumer {
// 	return &KafkaUserFollowConsumer{
// 		reader: kafka.NewReader(kafka.ReaderConfig{
// 			Brokers: []string{"localhost:9092"},
// 			Topic:   "follow",
// 			GroupID: "user-follow",
// 		}),
// 	}
// }

// func (k *KafkaUserFollowConsumer) ReadMessage(ctx context.Context) (event.FollowEvent, error) {
// 	m, err := k.reader.ReadMessage(ctx)
// 	m.
// 	if err != nil {
// 		return event.FollowEvent{}, fmt.Errorf("KafkaUserFollowConsumer.ReadMessage: %w", err)
// 	}

// 	var follow event.FollowEvent
// 	if err := json.Unmarshal(m.Value, &follow); err != nil {
// 		return event.FollowEvent{}, fmt.Errorf("KafkaUserFollowConsumer.ReadMessage: %w", err)
// 	}

// 	return follow, nil
// }
