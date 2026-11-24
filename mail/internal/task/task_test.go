package task

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"valid", `{"type":"otp","to":"a@b.com","otp":"123456","expires_at":"2026-11-24T10:00:00Z"}`, false},
		{"unknown fields are ignored", `{"type":"otp","to":"a@b.com","otp":"123456","expires_at":"2026-11-24T10:00:00Z","x":1}`, false},
		{"malformed json", `{"type":`, true},
		{"unknown type", `{"type":"welcome","to":"a@b.com","otp":"123456","expires_at":"2026-11-24T10:00:00Z"}`, true},
		{"missing to", `{"type":"otp","otp":"123456","expires_at":"2026-11-24T10:00:00Z"}`, true},
		{"not an address", `{"type":"otp","to":"nope","otp":"123456","expires_at":"2026-11-24T10:00:00Z"}`, true},
		{"display name", `{"type":"otp","to":"A <a@b.com>","otp":"123456","expires_at":"2026-11-24T10:00:00Z"}`, true},
		{"header injection", `{"type":"otp","to":"a@b.com\r\nBcc: x@y.com","otp":"123456","expires_at":"2026-11-24T10:00:00Z"}`, true},
		{"empty otp", `{"type":"otp","to":"a@b.com","otp":"","expires_at":"2026-11-24T10:00:00Z"}`, true},
		{"non digit otp", `{"type":"otp","to":"a@b.com","otp":"12ab56","expires_at":"2026-11-24T10:00:00Z"}`, true},
		{"too long otp", `{"type":"otp","to":"a@b.com","otp":"1234567890123","expires_at":"2026-11-24T10:00:00Z"}`, true},
		{"missing expires_at", `{"type":"otp","to":"a@b.com","otp":"123456"}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode([]byte(tt.body))
			if tt.wantErr {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("err = %v, want ErrInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
		})
	}
}

func TestDecode_ErrorHidesContent(t *testing.T) {
	_, err := Decode([]byte(`{"type":"otp","to":"secret@b.com","otp":"9x9","expires_at":"2026-11-24T10:00:00Z"}`))
	if err == nil {
		t.Fatal("want error")
	}
	for _, s := range []string{"secret@b.com", "9x9"} {
		if strings.Contains(err.Error(), s) {
			t.Fatalf("error %q leaks %q", err, s)
		}
	}
}

func TestExpired(t *testing.T) {
	exp := time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)
	task := Task{ExpiresAt: exp}
	if task.Expired(exp.Add(-time.Second)) {
		t.Error("expired one second before expires_at")
	}
	if !task.Expired(exp) {
		t.Error("not expired at expires_at")
	}
}
