package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/htan06/Moments/mail/internal/mailer"
)

var now = time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)

type fakeSender struct {
	err  error
	sent []mailer.Message
}

func (f *fakeSender) Send(_ context.Context, msg mailer.Message) error {
	f.sent = append(f.sent, msg)
	return f.err
}

func newHandler(send Sender, logs *bytes.Buffer) *Handler {
	h := New(send, slog.New(slog.NewTextHandler(logs, nil)))
	h.now = func() time.Time { return now }
	return h
}

func otpBody(expiresAt time.Time) []byte {
	return fmt.Appendf(nil, `{"type":"otp","to":"secret@b.com","otp":"482913","expires_at":%q}`,
		expiresAt.Format(time.RFC3339))
}

func TestHandle_Sent(t *testing.T) {
	var logs bytes.Buffer
	send := &fakeSender{}
	got := newHandler(send, &logs).Handle(t.Context(), "m1", otpBody(now.Add(10*time.Minute)))

	if got != Ack {
		t.Fatalf("outcome = %v, want Ack", got)
	}
	if len(send.sent) != 1 {
		t.Fatalf("sent %d mails, want 1", len(send.sent))
	}
	msg := send.sent[0]
	if msg.To != "secret@b.com" || !strings.Contains(msg.HTML, "482913") || !strings.Contains(msg.Text, "10 minutes") {
		t.Errorf("unexpected message: to=%q", msg.To)
	}
	assertNoSecrets(t, logs.String())
}

func TestHandle_ExpiredNotSent(t *testing.T) {
	var logs bytes.Buffer
	send := &fakeSender{}
	got := newHandler(send, &logs).Handle(t.Context(), "m1", otpBody(now))

	if got != Ack {
		t.Fatalf("outcome = %v, want Ack", got)
	}
	if len(send.sent) != 0 {
		t.Fatal("expired task was sent")
	}
}

func TestHandle_InvalidRejected(t *testing.T) {
	var logs bytes.Buffer
	send := &fakeSender{}
	got := newHandler(send, &logs).Handle(t.Context(), "m1",
		[]byte(`{"type":"otp","to":"secret@b.com","otp":"482913"}`))

	if got != Reject {
		t.Fatalf("outcome = %v, want Reject", got)
	}
	if len(send.sent) != 0 {
		t.Fatal("invalid task was sent")
	}
	assertNoSecrets(t, logs.String())
}

func TestHandle_TransientFailureRetried(t *testing.T) {
	var logs bytes.Buffer
	send := &fakeSender{err: errors.New("dial tcp: i/o timeout")}
	got := newHandler(send, &logs).Handle(t.Context(), "m1", otpBody(now.Add(time.Minute)))

	if got != Retry {
		t.Fatalf("outcome = %v, want Retry", got)
	}
}

func TestHandle_RecipientRefusedRejected(t *testing.T) {
	var logs bytes.Buffer
	send := &fakeSender{err: &mailer.SMTPError{Stage: "rcpt", Code: 550}}
	got := newHandler(send, &logs).Handle(t.Context(), "m1", otpBody(now.Add(time.Minute)))

	if got != Reject {
		t.Fatalf("outcome = %v, want Reject", got)
	}
}

func TestHandle_AuthFailureRetried(t *testing.T) {
	var logs bytes.Buffer
	send := &fakeSender{err: &mailer.SMTPError{Stage: "auth", Code: 535}}
	got := newHandler(send, &logs).Handle(t.Context(), "m1", otpBody(now.Add(time.Minute)))

	if got != Retry {
		t.Fatalf("outcome = %v, want Retry", got)
	}
}

func assertNoSecrets(t *testing.T, logs string) {
	t.Helper()
	for _, s := range []string{"secret@b.com", "482913"} {
		if strings.Contains(logs, s) {
			t.Errorf("logs leak %q: %s", s, logs)
		}
	}
}
