// Package mq declares every RabbitMQ message the user module publishes.
// Exchange and routing-key names come from config; topology is created by the
// dev, never by the app.
package mq

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	platformmq "github.com/htan06/Moments/application/internal/platform/mq"
)

// MailTask is the message consumed by the mail service. Its content (it holds
// the OTP) must never be logged.
type MailTask struct {
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	ExpiresAt time.Time `json:"expires_at"` // mail service drops the task after this
}

// MailRoute is where mail tasks are published.
type MailRoute struct {
	Exchange   string
	RoutingKey string
}

type MailPublisher struct {
	pub   *platformmq.Publisher
	route MailRoute
}

func NewMailPublisher(pub *platformmq.Publisher, route MailRoute) *MailPublisher {
	return &MailPublisher{pub: pub, route: route}
}

// SendOTP publishes an OTP email task. Template is temporary (see CLAUDE.md).
func (m *MailPublisher) SendOTP(ctx context.Context, email, otp string, expiresAt time.Time) error {
	body, err := json.Marshal(MailTask{
		To:        email,
		Subject:   "Moments verification code",
		Body:      "Your OTP is " + otp,
		ExpiresAt: expiresAt.UTC(),
	})
	if err != nil {
		return fmt.Errorf("mq: encode mail task: %w", err)
	}
	return m.pub.Publish(ctx, m.route.Exchange, m.route.RoutingKey, platformmq.Message{
		Body:      body,
		MessageID: messageID(),
	})
}

func messageID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b) // crypto/rand.Read never fails on supported platforms
	return hex.EncodeToString(b)
}
