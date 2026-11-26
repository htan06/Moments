package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestParseEmail(t *testing.T) {
	tests := []struct {
		in, want string
		err      error
	}{
		{"Alice@Example.com", "alice@example.com", nil},
		{"  bob@example.com ", "bob@example.com", nil},
		{"", "", ErrInvalidEmail},
		{"no-at-sign", "", ErrInvalidEmail},
		{"Alice <alice@example.com>", "", ErrInvalidEmail},
		{"a@b@c.com", "", ErrInvalidEmail},
		{"first.last+tag@mail.example.co", "first.last+tag@mail.example.co", nil},
		{"no-tld@localhost", "", ErrInvalidEmail},
		{"a@-bad.com", "", ErrInvalidEmail},
		{"a@bad-.com", "", ErrInvalidEmail},
		{"a@b..com", "", ErrInvalidEmail},
		{"sp ace@x.com", "", ErrInvalidEmail},
		{"tân@x.com", "", ErrInvalidEmail},
		{strings.Repeat("a", 250) + "@x.com", "", ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseEmail(tt.in)
			if !errors.Is(err, tt.err) || got != tt.want {
				t.Fatalf("ParseEmail(%q) = %q, %v; want %q, %v", tt.in, got, err, tt.want, tt.err)
			}
		})
	}
}

func TestUsernameRules_Parse(t *testing.T) {
	r := NewUsernameRules([]string{"Admin", "moments"})
	tests := []struct {
		name, in, want string
		err            error
	}{
		{"lowercased", "Alice_01", "alice_01", nil},
		{"dot inside", "a.b.c", "a.b.c", nil},
		{"min length", "abc", "abc", nil},
		{"max length", strings.Repeat("a", 30), strings.Repeat("a", 30), nil},
		{"too short", "ab", "", ErrInvalidUsername},
		{"too long", strings.Repeat("a", 31), "", ErrInvalidUsername},
		{"leading dot", ".abc", "", ErrInvalidUsername},
		{"trailing dot", "abc.", "", ErrInvalidUsername},
		{"double dot", "a..b", "", ErrInvalidUsername},
		{"at sign", "a@bc", "", ErrInvalidUsername},
		{"dash", "a-bc", "", ErrInvalidUsername},
		{"unicode", "tân123", "", ErrInvalidUsername},
		{"reserved any case", "ADMIN", "", ErrInvalidUsername},
		{"reserved", "moments", "", ErrInvalidUsername},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Parse(tt.in)
			if !errors.Is(err, tt.err) || got != tt.want {
				t.Fatalf("Parse(%q) = %q, %v; want %q, %v", tt.in, got, err, tt.want, tt.err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name, in string
		ok       bool
	}{
		{"valid", "abc123", true},
		{"unicode letters", "mậtkhẩu1", true},
		{"max length", strings.Repeat("a", 63) + "1", true},
		{"too short", "ab12", false},
		{"too long", strings.Repeat("a", 64) + "1", false},
		{"no digit", "abcdefg", false},
		{"no letter", "1234567", false},
		{"invalid utf8", "abc12\xff", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.in)
			if (err == nil) != tt.ok {
				t.Fatalf("ValidatePassword(%q) = %v, want ok=%v", tt.in, err, tt.ok)
			}
		})
	}
}

func TestNewAccount(t *testing.T) {
	a, err := NewAccount("a@x.com", "alice", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if a.Status != StatusActive || !a.Active() {
		t.Errorf("status = %q, want ACTIVE", a.Status)
	}
	if _, err := NewAccount("a@x.com", "alice", ""); !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("empty hash: err = %v, want ErrInvalidPassword", err)
	}
}

func TestIsEmailIdentifier(t *testing.T) {
	if !IsEmailIdentifier("a@x.com") || IsEmailIdentifier("alice") {
		t.Fatal("IsEmailIdentifier misclassified")
	}
}

func TestSetPasswordHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr error
	}{
		{"new hash", "hash:new", nil},
		{"empty hash", "", ErrInvalidPassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{PasswordHash: "hash:old"}
			err := a.SetPasswordHash(tt.hash)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			want := tt.hash
			if tt.wantErr != nil {
				want = "hash:old"
			}
			if a.PasswordHash != want {
				t.Fatalf("hash = %q, want %q", a.PasswordHash, want)
			}
		})
	}
}

func TestRename(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		wantChanged bool
		wantErr     error
		wantName    string
	}{
		{"new username", "bob", true, nil, "bob"},
		{"same username", "alice", false, nil, "alice"},
		{"empty", "", false, ErrInvalidUsername, "alice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{Username: "alice"}
			changed, err := a.Rename(tt.in)
			if changed != tt.wantChanged || !errors.Is(err, tt.wantErr) || a.Username != tt.wantName {
				t.Fatalf("Rename = %v, %v, username %q; want %v, %v, %q", changed, err, a.Username, tt.wantChanged, tt.wantErr, tt.wantName)
			}
		})
	}
}
