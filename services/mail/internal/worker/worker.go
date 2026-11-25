// Package worker decides what happens to one mail task: send it, drop it, or
// put it back for a retry. It knows nothing about RabbitMQ.
package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/htan06/Moments/mail/internal/compose"
	"github.com/htan06/Moments/mail/internal/mailer"
	"github.com/htan06/Moments/mail/internal/task"
)

// Outcome tells the consumer how to settle the delivery.
type Outcome int

const (
	Ack    Outcome = iota // done: sent, or dropped on purpose (expired)
	Reject                // can never succeed: drop without requeue
	Retry                 // transient failure: requeue
)

type Sender interface {
	Send(ctx context.Context, msg mailer.Message) error
}

type Handler struct {
	send Sender
	now  func() time.Time
	log  *slog.Logger
}

func New(send Sender, log *slog.Logger) *Handler {
	return &Handler{send: send, now: time.Now, log: log}
}

// Handle processes one task body. id is the AMQP message id: the only task
// detail ever logged (never the address, the OTP or the body).
func (h *Handler) Handle(ctx context.Context, id string, body []byte) Outcome {
	log := h.log.With(slog.String("message_id", id))

	t, err := task.Decode(body)
	if err != nil {
		log.Warn("mail task rejected", slog.String("reason", err.Error()))
		return Reject
	}
	now := h.now()
	if t.Expired(now) {
		// USER-R2: the OTP is useless now; the user asks for a new one.
		log.Info("mail task expired, not sent", slog.String("type", t.Type))
		return Ack
	}

	email, err := compose.OTP(t.OTP, t.ExpiresAt.Sub(now))
	if err != nil {
		log.Error("render mail", slog.Any("err", err))
		return Reject
	}
	err = h.send.Send(ctx, mailer.Message{
		To:      t.To,
		Subject: email.Subject,
		Text:    email.Text,
		HTML:    email.HTML,
	})
	switch {
	case err == nil:
		log.Info("mail sent", slog.String("type", t.Type))
		return Ack
	case mailer.IsRejected(err):
		log.Warn("mail refused by smtp server", slog.Any("err", err))
		return Reject
	default:
		log.Error("send mail, will retry", slog.Any("err", err))
		return Retry
	}
}
