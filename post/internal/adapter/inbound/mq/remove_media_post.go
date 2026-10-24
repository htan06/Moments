package messagequeue

// import (
// 	"context"
// 	"encoding/json"
// 	"log"

// 	"github.com/htan06/Moments/post/internal/domain"
// 	"github.com/minio/minio-go/v7"
// 	"github.com/segmentio/kafka-go"
// )

// type RemoveMediaWorker struct {
// 	reader    *kafka.Reader
// 	minioConn *minio.Client
// }

// func NewKafkaPostConsumer(minioConn *minio.Client) *RemoveMediaWorker {
// 	return &RemoveMediaWorker{
// 		reader: kafka.NewReader(kafka.ReaderConfig{
// 			Brokers: []string{"localhost:9092"},
// 			Topic:   "post",
// 			GroupID: "remove-media-post",
// 		}),
// 		minioConn: minioConn,
// 	}
// }

// func (r *RemoveMediaWorker) Run(ctx context.Context) {
// 	for {
// 		msg, err := r.reader.FetchMessage(ctx)
// 		if err != nil {
// 			log.Println("Error: %s", err.Error())
// 		}

// 		var event domain.PostEvent
// 		if err := json.Unmarshal(msg.Value, &event); err != nil {
// 			log.Println("Error: %s", err.Error())
// 		}

// 		r.minioConn.Remove(ctx, "posts", )
// 	}
// }
