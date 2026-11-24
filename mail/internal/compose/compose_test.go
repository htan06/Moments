package compose

import (
	"strings"
	"testing"
	"time"
)

func TestOTP_ContainsCodeAndExpiry(t *testing.T) {
	e, err := OTP("482913", 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if e.Subject == "" {
		t.Error("empty subject")
	}
	for name, body := range map[string]string{"text": e.Text, "html": e.HTML} {
		if !strings.Contains(body, "482913") {
			t.Errorf("%s body has no code", name)
		}
		if !strings.Contains(body, "10 minutes") {
			t.Errorf("%s body has no expiry", name)
		}
	}
}

func TestMinutes(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{10 * time.Minute, "10 minutes"},
		{9*time.Minute + 58*time.Second, "10 minutes"},
		{9 * time.Minute, "9 minutes"},
		{61 * time.Second, "2 minutes"},
		{time.Minute, "1 minute"},
		{30 * time.Second, "1 minute"},
	}
	for _, tt := range tests {
		if got := minutes(tt.in); got != tt.want {
			t.Errorf("minutes(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
