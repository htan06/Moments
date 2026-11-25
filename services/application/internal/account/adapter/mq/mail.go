// Package mq declares every RabbitMQ message the account module publishes.
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

// mailTypeOTP is MailTask.Type for an OTP email.
const mailTypeOTP = "otp"

// MailTask is the message consumed by the mail service, which renders subject
// and body itself. Its content (it holds the OTP) must never be logged.
type MailTask struct {
	Type      string    `json:"type"`       // "otp"
	To        string    `json:"to"`         // bare address: "a@b.com", no display name
	OTP       string    `json:"otp"`        // digits only
	ExpiresAt time.Time `json:"expires_at"` // RFC 3339, UTC; mail service drops the task after this
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

// SendOTP publishes an OTP email task. email is already a bare, lowercased
// address (domain.ParseEmail).
func (m *MailPublisher) SendOTP(ctx context.Context, email, otp string, expiresAt time.Time) error {
	body, err := json.Marshal(MailTask{
		Type:      mailTypeOTP,
		To:        email,
		OTP:       otp,
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
