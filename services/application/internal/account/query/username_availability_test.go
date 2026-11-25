package query

import (
	"context"
	"errors"
	"testing"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// fakeUsers records the username it was asked about.
type fakeUsers struct {
	taken map[string]bool
	asked string
}

func (f *fakeUsers) UsernameExists(_ context.Context, username string) (bool, error) {
	f.asked = username
	return f.taken[username], nil
}

func newHandler() (*UsernameAvailabilityHandler, *fakeUsers) {
	users := &fakeUsers{taken: map[string]bool{"alice": true}}
	return NewUsernameAvailabilityHandler(users, domain.NewUsernameRules([]string{"admin"})), users
}

func TestUsernameAvailability_Taken(t *testing.T) {
	h, users := newHandler()
	ok, err := h.Handle(context.Background(), UsernameAvailability{Username: " ALICE "})
	if err != nil || ok {
		t.Fatalf("Handle = %v, %v; want false, nil", ok, err)
	}
	if users.asked != "alice" {
		t.Fatalf("asked %q, want normalized alice", users.asked)
	}
}

func TestUsernameAvailability_Free(t *testing.T) {
	h, _ := newHandler()
	if ok, err := h.Handle(context.Background(), UsernameAvailability{Username: "bob"}); err != nil || !ok {
		t.Fatalf("Handle = %v, %v; want true, nil", ok, err)
	}
}

func TestUsernameAvailability_InvalidOrReserved(t *testing.T) {
	h, users := newHandler()
	for _, name := range []string{"a", "bad..name", "admin"} {
		if _, err := h.Handle(context.Background(), UsernameAvailability{Username: name}); !errors.Is(err, domain.ErrInvalidUsername) {
			t.Fatalf("%q: err = %v, want ErrInvalidUsername", name, err)
		}
	}
	if users.asked != "" {
		t.Fatal("invalid username reached the database")
	}
}
