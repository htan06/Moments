package command

import (
	"context"
	"errors"
	"testing"

	"github.com/htan06/Moments/application/internal/user/domain"
)

// fakeUsernames mimics the profiles table: one username copy per user, unique.
type fakeUsernames struct {
	m   map[int64]string
	err error
}

func (f *fakeUsernames) UpdateUsername(_ context.Context, userID int64, username string) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.m[userID]; !ok {
		return domain.ErrProfileNotFound
	}
	for id, u := range f.m {
		if u == username && id != userID {
			return domain.ErrUsernameTaken
		}
	}
	f.m[userID] = username
	return nil
}

func TestUpdateUsername(t *testing.T) {
	tests := []struct {
		name     string
		userID   int64
		username string
		want     error
		wantCopy string // username copy of user 1 afterwards
	}{
		{"renames", 1, "bob", nil, "bob"},
		{"same username is idempotent", 1, "alice", nil, "alice"},
		{"username held by another profile", 1, "carol", domain.ErrUsernameTaken, "alice"},
		{"no profile", 9, "bob", domain.ErrProfileNotFound, "alice"},
		{"empty username", 1, "", domain.ErrInvalidUsername, "alice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeUsernames{m: map[int64]string{1: "alice", 2: "carol"}}
			err := NewUpdateUsernameHandler(store).Handle(context.Background(), UpdateUsername{UserID: tt.userID, Username: tt.username})
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if got := store.m[1]; got != tt.wantCopy {
				t.Fatalf("copy = %q, want %q", got, tt.wantCopy)
			}
		})
	}
}

func TestUpdateUsername_StoreErrorIsWrapped(t *testing.T) {
	store := &fakeUsernames{m: map[int64]string{1: "alice"}, err: errors.New("db down")}
	err := NewUpdateUsernameHandler(store).Handle(context.Background(), UpdateUsername{UserID: 1, Username: "bob"})
	if err == nil || errors.Is(err, domain.ErrUsernameTaken) || errors.Is(err, domain.ErrProfileNotFound) {
		t.Fatalf("err = %v, want a wrapped store error", err)
	}
}
