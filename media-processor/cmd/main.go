package main

import (
	"context"
	"log"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/htan06/Moments/media-processor/internal/config"
	"github.com/htan06/Moments/media-processor/internal/infra"
	"github.com/htan06/Moments/media-processor/internal/orchestrator"
	"github.com/htan06/Moments/media-processor/internal/processor"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("WAR: Cannot load .env file")
	}

	if err := vips.Startup(nil); err != nil {
		log.Fatalln(err)
	}
	defer vips.Shutdown()

	minioConn := config.GetMinIOConn()
	postgresConn := config.GetPostgresConn()

	mediaRepo := infra.NewPostgresMediarepository(postgresConn)

	imageProcessor := processor.NewImageProcessor(minioConn)
	videoProcessor := processor.NewVideoProcessor(minioConn)

	orchestrator := orchestrator.NewOrchestrator(
		minioConn,
		imageProcessor,
		videoProcessor,
		mediaRepo,
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "media-job",
			GroupID: "media-processor",
		}))

	orchestrator.Start(context.Background())
}
