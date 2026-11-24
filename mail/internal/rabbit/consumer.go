// Package rabbit consumes mail tasks from RabbitMQ. The queue and its binding
// to the mail exchange are created by the dev; this package never declares
// topology.
package rabbit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/htan06/Moments/mail/internal/worker"
)

type Handler interface {
	Handle(ctx context.Context, id string, body []byte) worker.Outcome
}

type Config struct {
	URL            string // contains credentials: never log
	Queue          string
	Workers        int           // tasks handled in parallel; also the prefetch
	RetryDelay     time.Duration // wait before requeueing a failed task
	ReconnectDelay time.Duration // wait before reconnecting after a drop
}

type Consumer struct {
	cfg Config
	h   Handler
	log *slog.Logger
}

func NewConsumer(cfg Config, h Handler, log *slog.Logger) (*Consumer, error) {
	if cfg.Workers <= 0 {
		return nil, fmt.Errorf("rabbit: workers must be > 0, got %d", cfg.Workers)
	}
	return &Consumer{cfg: cfg, h: h, log: log}, nil
}

// Run consumes until ctx is done, reconnecting when the connection or channel
// drops. If the very first connection fails (wrong credentials, missing
// queue), Run returns the error so misconfiguration fails at startup.
// Tasks already received when ctx ends are finished and settled first.
func (c *Consumer) Run(ctx context.Context) error {
	for first := true; ; first = false {
		started, err := c.session(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if first && !started {
			return err
		}
		c.log.Error("rabbitmq consumer stopped, reconnecting",
			slog.Any("err", err), slog.Duration("in", c.cfg.ReconnectDelay))
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.cfg.ReconnectDelay):
		}
	}
}

// session runs one connection until it breaks or ctx ends. started reports
// whether consuming began.
func (c *Consumer) session(ctx context.Context) (started bool, err error) {
	conn, err := amqp.Dial(c.cfg.URL)
	if err != nil {
		// amqp errors do not include the URL, safe to wrap.
		return false, fmt.Errorf("rabbit: dial: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return false, fmt.Errorf("rabbit: open channel: %w", err)
	}
	// Prefetch = workers: the broker never hands this process more unacked
	// tasks than it can work on, the rest stay in the queue for other replicas.
	if err := ch.Qos(c.cfg.Workers, 0, false); err != nil {
		return false, fmt.Errorf("rabbit: qos: %w", err)
	}
	tag := consumerTag()
	deliveries, err := ch.Consume(c.cfg.Queue, tag, false, false, false, false, nil)
	if err != nil {
		return false, fmt.Errorf("rabbit: consume %q: %w", c.cfg.Queue, err)
	}
	c.log.Info("consuming mail tasks", slog.String("queue", c.cfg.Queue), slog.Int("workers", c.cfg.Workers))

	// Sends keep a context that is not cancelled on shutdown, so a mail that
	// is half sent finishes and gets acked (the SMTP timeout still bounds it).
	work := context.WithoutCancel(ctx)
	var wg sync.WaitGroup
	for range c.cfg.Workers {
		wg.Go(func() {
			for d := range deliveries {
				c.handle(ctx, work, d)
			}
		})
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		// The delivery channel closes when the connection or channel drops.
		return true, errors.New("rabbit: delivery channel closed")
	case <-ctx.Done():
		// Stop new deliveries; workers drain what was already received.
		// Prefetched tasks not handled yet are requeued by the broker when
		// the channel closes.
		if err := ch.Cancel(tag, false); err != nil {
			c.log.Warn("cancel consumer", slog.Any("err", err))
		}
		<-done
		return true, nil
	}
}

func (c *Consumer) handle(ctx, work context.Context, d amqp.Delivery) {
	var err error
	switch c.h.Handle(work, d.MessageId, d.Body) {
	case worker.Ack:
		err = d.Ack(false)
	case worker.Reject:
		// Goes to the queue's dead-letter exchange if the dev configured one,
		// otherwise it is dropped.
		err = d.Reject(false)
	case worker.Retry:
		// Requeued tasks come back almost at once: wait first, so a failing
		// SMTP server is not hit in a tight loop. The task's expires_at caps
		// the retries: once expired, the next delivery drops it.
		select {
		case <-ctx.Done():
		case <-time.After(c.cfg.RetryDelay):
		}
		err = d.Nack(false, true)
	}
	if err != nil {
		c.log.Warn("settle delivery", slog.String("message_id", d.MessageId), slog.Any("err", err))
	}
}

func consumerTag() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read never fails on supported platforms
	return "mail-" + hex.EncodeToString(b)
}
