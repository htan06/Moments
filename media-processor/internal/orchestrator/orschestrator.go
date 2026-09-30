package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/htan06/Moments/media-processor/internal/domain"
	"github.com/htan06/Moments/media-processor/internal/processor"
	"github.com/minio/minio-go/v7"
	"github.com/segmentio/kafka-go"
)

type Orchestrator struct {
	minioConn      *minio.Client
	imageProcessor *processor.ImageProcessor
	videoProcessor *processor.VideoProcessor
	mediaRepo      domain.MediaRepository
	kafkaReader    *kafka.Reader
}

func NewOrchestrator(
	minioConn *minio.Client,
	imageProcessor *processor.ImageProcessor,
	videoProcessor *processor.VideoProcessor,
	mediaRepo domain.MediaRepository,
	kafkaReader *kafka.Reader,
) *Orchestrator {
	return &Orchestrator{
		minioConn:      minioConn,
		imageProcessor: imageProcessor,
		videoProcessor: videoProcessor,
		mediaRepo:      mediaRepo,
		kafkaReader:    kafkaReader,
	}
}

func (d *Orchestrator) Start(ctx context.Context) {
	for {
		msg, err := d.kafkaReader.FetchMessage(ctx)
		if err != nil {
			log.Println(err)
		}

		job, err := extractJob(msg)
		if err != nil {
			log.Println(err)
		}

		if job.MediaType == domain.Image {
			if err := d.imageProcessor.ProcessJob(ctx, job); err != nil {
				log.Println(err)
			}
		} else if job.MediaType == domain.Video {
			if err := d.videoProcessor.ProcessJob(ctx, job); err != nil {
				log.Println(err)
			}
		}
		if err := d.kafkaReader.CommitMessages(ctx, msg); err != nil {
			log.Fatalln(err)
		}
	}
}

func extractJob(msg kafka.Message) (domain.ProcessMediaJob, error) {
	job := domain.ProcessMediaJob{}
	if err := json.Unmarshal(msg.Value, &job); err != nil {
		return domain.ProcessMediaJob{}, fmt.Errorf("extractJob: %w", err)
	}
	return job, nil
}
