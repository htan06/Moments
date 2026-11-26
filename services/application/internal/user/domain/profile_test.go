package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestParseName(t *testing.T) {
	tests := []struct {
		name, in, want string
		err            error
	}{
		{"trimmed", "  Alice Nguyễn ", "Alice Nguyễn", nil},
		{"max length counts runes", strings.Repeat("ễ", 50), strings.Repeat("ễ", 50), nil},
		{"blank", "   ", "", ErrInvalidName},
		{"too long", strings.Repeat("a", 51), "", ErrInvalidName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseName(tt.in)
			if !errors.Is(err, tt.err) || got != tt.want {
				t.Fatalf("ParseName(%q) = %q, %v; want %q, %v", tt.in, got, err, tt.want, tt.err)
			}
		})
	}
}

func TestNewProfile(t *testing.T) {
	p, err := NewProfile(7, "alice", " Alice ")
	if err != nil || p != (Profile{UserID: 7, Username: "alice", Name: "Alice"}) {
		t.Fatalf("NewProfile = %+v, %v", p, err)
	}
	if _, err := NewProfile(7, "alice", ""); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("err = %v, want ErrInvalidName", err)
	}
	if _, err := NewProfile(7, "", "Alice"); !errors.Is(err, ErrInvalidUsername) {
		t.Fatalf("err = %v, want ErrInvalidUsername", err)
	}
}
