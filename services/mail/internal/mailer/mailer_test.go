package mailer

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"
	"time"
)

func TestBuild_ParsesAsAlternative(t *testing.T) {
	from := &mail.Address{Name: "Moments", Address: "moments.app@gmail.com"}
	raw, err := build(from, Message{
		To:      "a@b.com",
		Subject: "Mã xác thực",
		Text:    "Your code: 123456",
		HTML:    "<p>" + strings.Repeat("x", 2000) + "</p>",
	}, time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	m, err := mail.ReadMessage(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	if got := m.Header.Get("To"); got != "a@b.com" {
		t.Errorf("To = %q", got)
	}
	subject, err := new(mime.WordDecoder).DecodeHeader(m.Header.Get("Subject"))
	if err != nil || subject != "Mã xác thực" {
		t.Errorf("Subject = %q, err %v", subject, err)
	}

	mediaType, params, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/alternative" {
		t.Fatalf("Content-Type = %q, err %v", mediaType, err)
	}
	mr := multipart.NewReader(m.Body, params["boundary"])
	var types []string
	for {
		p, err := mr.NextPart() // decodes quoted-printable
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(p)
		types = append(types, p.Header.Get("Content-Type"))
		if strings.HasPrefix(p.Header.Get("Content-Type"), "text/html") && len(body) != len("<p></p>")+2000 {
			t.Errorf("html part length %d after decoding", len(body))
		}
	}
	if len(types) != 2 || !strings.HasPrefix(types[0], "text/plain") || !strings.HasPrefix(types[1], "text/html") {
		t.Errorf("parts = %v", types)
	}

	for i, line := range strings.Split(string(raw), "\r\n") {
		if len(line) > 998 {
			t.Fatalf("line %d is %d bytes, over the SMTP limit", i, len(line))
		}
	}
}

func TestWrap_HidesServerText(t *testing.T) {
	err := wrap("rcpt", &textproto.Error{Code: 550, Msg: "5.1.1 secret@b.com does not exist"})
	if strings.Contains(err.Error(), "secret@b.com") {
		t.Fatalf("error leaks the address: %v", err)
	}
}

func TestIsRejected(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"unknown mailbox", wrap("rcpt", &textproto.Error{Code: 550}), true},
		{"content refused", wrap("data", &textproto.Error{Code: 552}), true},
		{"mailbox busy", wrap("rcpt", &textproto.Error{Code: 450}), false},
		{"wrong app password", wrap("auth", &textproto.Error{Code: 535}), false},
		{"network", wrap("data", errors.New("connection reset")), false},
		{"wrapped again", fmt.Errorf("send: %w", wrap("rcpt", &textproto.Error{Code: 550})), true},
	}
	for _, tt := range tests {
		if got := IsRejected(tt.err); got != tt.want {
			t.Errorf("%s: IsRejected = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSend_RejectsLineBreakInRecipient(t *testing.T) {
	m, err := New(Config{Host: "smtp.example.com", Port: "587", From: "a@example.com", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Send(t.Context(), Message{To: "a@b.com\r\nBcc: x@y.com"}); err == nil {
		t.Fatal("want error")
	}
}
