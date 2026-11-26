package command

import (
	"context"
	"errors"
	"testing"

	"github.com/htan06/Moments/application/internal/user/domain"
)

// fakeProfiles keeps the first profile stored per user, like the idempotent
// insert of the real store.
type fakeProfiles struct {
	m   map[int64]domain.Profile
	err error
}

func (f *fakeProfiles) Create(_ context.Context, p domain.Profile) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.m[p.UserID]; !ok {
		f.m[p.UserID] = p
	}
	return nil
}

func TestCreateProfile_StoresTrimmedName(t *testing.T) {
	store := &fakeProfiles{m: map[int64]domain.Profile{}}
	h := NewCreateProfileHandler(store)
	if err := h.Handle(context.Background(), CreateProfile{UserID: 7, Username: "alice", Name: " Alice "}); err != nil {
		t.Fatal(err)
	}
	if got := store.m[7]; got != (domain.Profile{UserID: 7, Username: "alice", Name: "Alice"}) {
		t.Fatalf("profile = %+v", got)
	}
}

func TestCreateProfile_RetryIsIdempotent(t *testing.T) {
	store := &fakeProfiles{m: map[int64]domain.Profile{}}
	h := NewCreateProfileHandler(store)
	for range 2 {
		if err := h.Handle(context.Background(), CreateProfile{UserID: 7, Username: "alice", Name: "Alice"}); err != nil {
			t.Fatal(err)
		}
	}
	if len(store.m) != 1 {
		t.Fatalf("profiles = %d, want 1", len(store.m))
	}
}

func TestCreateProfile_UsernameTaken(t *testing.T) {
	store := &fakeProfiles{m: map[int64]domain.Profile{}, err: domain.ErrUsernameTaken}
	h := NewCreateProfileHandler(store)
	if err := h.Handle(context.Background(), CreateProfile{UserID: 7, Username: "alice", Name: "Alice"}); !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("err = %v, want ErrUsernameTaken", err)
	}
}

func TestCreateProfile_InvalidName(t *testing.T) {
	store := &fakeProfiles{m: map[int64]domain.Profile{}}
	h := NewCreateProfileHandler(store)
	if err := h.Handle(context.Background(), CreateProfile{UserID: 7, Username: "alice", Name: "  "}); !errors.Is(err, domain.ErrInvalidName) {
		t.Fatalf("err = %v, want ErrInvalidName", err)
	}
	if len(store.m) != 0 {
		t.Fatal("invalid profile stored")
	}
}
