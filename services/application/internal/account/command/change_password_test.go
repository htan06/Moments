package command

import (
	"context"
	"errors"
	"testing"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// alice (login_test.go) has password secret1.
func newChangePassword(accounts ...domain.Account) (*ChangePasswordHandler, *fakeAccounts) {
	store := newFakeAccounts(accounts...)
	return NewChangePasswordHandler(store, &fakePasswords{}), store
}

func TestChangePassword_Success(t *testing.T) {
	h, store := newChangePassword(alice)
	if err := h.Handle(context.Background(), ChangePassword{UserID: 1, OldPassword: "secret1", NewPassword: "new1pass"}); err != nil {
		t.Fatal(err)
	}
	if got := store.byID[1].PasswordHash; got != "hash:new1pass" {
		t.Fatalf("hash = %q", got)
	}
}

func TestChangePassword_Errors(t *testing.T) {
	inactive := alice
	inactive.Status = "BANNED"
	tests := []struct {
		name    string
		account domain.Account
		cmd     ChangePassword
		want    error
	}{
		{"wrong current password", alice, ChangePassword{UserID: 1, OldPassword: "nope1", NewPassword: "new1pass"}, domain.ErrWrongPassword},
		{"same as current password", alice, ChangePassword{UserID: 1, OldPassword: "secret1", NewPassword: "secret1"}, domain.ErrSamePassword},
		{"malformed new password", alice, ChangePassword{UserID: 1, OldPassword: "secret1", NewPassword: "short"}, domain.ErrInvalidPassword},
		{"account gone", alice, ChangePassword{UserID: 9, OldPassword: "secret1", NewPassword: "new1pass"}, domain.ErrAccountNotFound},
		{"account not active", inactive, ChangePassword{UserID: 1, OldPassword: "secret1", NewPassword: "new1pass"}, domain.ErrAccountInactive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, store := newChangePassword(tt.account)
			if err := h.Handle(context.Background(), tt.cmd); !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if got := store.byID[1].PasswordHash; got != "hash:secret1" {
				t.Fatalf("hash changed to %q", got)
			}
		})
	}
}

func TestChangePassword_WrongPasswordIsNotLimited(t *testing.T) {
	h, _ := newChangePassword(alice)
	ctx := context.Background()
	for range 10 {
		if err := h.Handle(ctx, ChangePassword{UserID: 1, OldPassword: "nope1", NewPassword: "new1pass"}); !errors.Is(err, domain.ErrWrongPassword) {
			t.Fatalf("err = %v, want ErrWrongPassword", err)
		}
	}
	if err := h.Handle(ctx, ChangePassword{UserID: 1, OldPassword: "secret1", NewPassword: "new1pass"}); err != nil {
		t.Fatalf("err = %v after wrong attempts, want nil", err)
	}
}

func TestChangePassword_StoreError(t *testing.T) {
	h, store := newChangePassword(alice)
	store.updateErr = errors.New("db down")
	if err := h.Handle(context.Background(), ChangePassword{UserID: 1, OldPassword: "secret1", NewPassword: "new1pass"}); err == nil {
		t.Fatal("want error")
	}
}
