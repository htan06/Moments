// Command mail runs the Moments mail service: it consumes mail tasks from
// RabbitMQ and sends them over SMTP.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/htan06/Moments/mail/config"
	"github.com/htan06/Moments/mail/internal/mailer"
	"github.com/htan06/Moments/mail/internal/rabbit"
	"github.com/htan06/Moments/mail/internal/worker"
)

const startupTimeout = 30 * time.Second

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("mail service stopped", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Local dev (go run from services/mail/): load .env (localhost addresses)
	// if present. Variables already set in the environment win; the container
	// gets its variables from .env.docker via compose.
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	m, err := mailer.New(mailer.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		Timeout:  cfg.Worker.SMTPTimeout,
	})
	if err != nil {
		return err
	}
	startCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	if err := m.Verify(startCtx); err != nil {
		return fmt.Errorf("smtp login check: %w", err)
	}

	consumer, err := rabbit.NewConsumer(rabbit.Config{
		URL:            cfg.RabbitMQ.URL(),
		Queue:          cfg.RabbitMQ.Queue,
		Workers:        cfg.Worker.Count,
		RetryDelay:     cfg.Worker.RetryDelay,
		ReconnectDelay: cfg.Worker.ReconnectDelay,
	}, worker.New(m, log), log)
	if err != nil {
		return err
	}

	log.Info("mail service started")
	if err := consumer.Run(ctx); err != nil {
		return err
	}
	log.Info("mail service stopped")
	return nil
}
