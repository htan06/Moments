// Package mq provides the RabbitMQ connection and a confirming publisher.
// Topology (exchanges, queues, bindings) is managed by the dev outside the
// app; this package never declares it.
package mq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Message is one message to publish. Body is never logged.
type Message struct {
	Body        []byte
	ContentType string // defaults to application/json
	MessageID   string
}

var (
	ErrNacked = errors.New("mq: message nacked by broker")
	ErrClosed = errors.New("mq: publisher closed")
)

// Publisher publishes persistent messages and waits for broker confirms.
//
// All channels are multiplexed over one TCP connection. The publisher keeps a
// pool of up to `channels` confirm-mode channels; each Publish borrows one, so
// up to `channels` publishes (and their confirm waits) run in parallel.
// Channels are opened lazily and reopened after errors; the connection is
// redialed if it drops.
type Publisher struct {
	url string

	connMu sync.Mutex
	conn   *amqp.Connection
	closed bool

	// slots holds one entry per allowed channel; nil means "not opened yet".
	slots chan *amqp.Channel
}

// NewPublisher connects eagerly so misconfiguration fails at startup.
// url contains credentials: never log it.
func NewPublisher(url string, channels int) (*Publisher, error) {
	if channels <= 0 {
		return nil, fmt.Errorf("mq: channels must be > 0, got %d", channels)
	}
	p := &Publisher{url: url, slots: make(chan *amqp.Channel, channels)}

	first, err := p.openChannel()
	if err != nil {
		return nil, err
	}
	p.slots <- first
	for i := 1; i < channels; i++ {
		p.slots <- nil
	}
	return p, nil
}

// Publish sends msg to exchange with routingKey and waits for the confirm.
// It blocks while all channels are busy, until ctx is done.
func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, msg Message) error {
	var ch *amqp.Channel
	select {
	case ch = <-p.slots:
	case <-ctx.Done():
		return fmt.Errorf("mq: wait for channel: %w", ctx.Err())
	}
	// Always give the slot back; a broken channel goes back as nil.
	defer func() { p.slots <- ch }()

	if ch == nil || ch.IsClosed() {
		var err error
		if ch, err = p.openChannel(); err != nil {
			ch = nil
			return err
		}
	}

	contentType := msg.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	dc, err := ch.PublishWithDeferredConfirmWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  contentType,
		MessageId:    msg.MessageID,
		Timestamp:    time.Now(),
		Body:         msg.Body,
	})
	if err != nil {
		_ = ch.Close()
		ch = nil
		return fmt.Errorf("mq: publish to %q/%q: %w", exchange, routingKey, err)
	}
	acked, err := dc.WaitContext(ctx)
	if err != nil {
		// Confirm state of this channel is unknown now; drop it.
		_ = ch.Close()
		ch = nil
		return fmt.Errorf("mq: wait confirm %q/%q: %w", exchange, routingKey, err)
	}
	if !acked {
		return fmt.Errorf("%w: %q/%q", ErrNacked, exchange, routingKey)
	}
	return nil
}

// Close closes the connection, which closes every channel.
func (p *Publisher) Close() error {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	p.closed = true
	if p.conn != nil && !p.conn.IsClosed() {
		return p.conn.Close()
	}
	return nil
}

// openChannel opens a confirm-mode channel, redialing the connection if needed.
func (p *Publisher) openChannel() (*amqp.Channel, error) {
	conn, err := p.connection()
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("mq: open channel: %w", err)
	}
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		return nil, fmt.Errorf("mq: enable confirms: %w", err)
	}
	return ch, nil
}

func (p *Publisher) connection() (*amqp.Connection, error) {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	if p.closed {
		return nil, ErrClosed
	}
	if p.conn == nil || p.conn.IsClosed() {
		conn, err := amqp.Dial(p.url)
		if err != nil {
			// amqp errors do not include the URL, safe to wrap.
			return nil, fmt.Errorf("mq: dial: %w", err)
		}
		p.conn = conn
	}
	return p.conn, nil
}
