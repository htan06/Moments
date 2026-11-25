package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/htan06/Moments/application/internal/platform/token"
	"github.com/htan06/Moments/application/internal/user/domain"
)

const dummyHash = "hash:dummy"

func newLogin(accounts ...domain.Account) (*LoginHandler, *fakePasswords, *fakeAttempts, *fakeTokens) {
	pw, at, tk := &fakePasswords{}, newFakeAttempts(), &fakeTokens{}
	return NewLoginHandler(newFakeAccounts(accounts...), pw, at, tk, dummyHash), pw, at, tk
}

var alice = activeAccount(1, "alice@example.com", "alice", "secret1")

func TestLogin_SuccessWithUsernameOrEmail(t *testing.T) {
	for _, id := range []string{"alice", "ALICE", "Alice@Example.com"} {
		h, _, _, tk := newLogin(alice)
		pair, err := h.Handle(context.Background(), Login{Identifier: id, Password: "secret1"})
		if err != nil || pair.AccessToken != "access-alice" {
			t.Fatalf("%s: Handle = %+v, %v", id, pair, err)
		}
		if tk.issued[0] != (token.Claims{UserID: 1, Username: "alice", Status: domain.StatusActive}) {
			t.Fatalf("claims = %+v", tk.issued[0])
		}
	}
}

func TestLogin_WrongPasswordCountsFailureOnAccount(t *testing.T) {
	h, _, at, _ := newLogin(alice)
	_, err := h.Handle(context.Background(), Login{Identifier: "alice", Password: "wrong1"})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
	if at.fails[LoginSubject{AccountID: 1}] != 1 {
		t.Fatalf("fails = %+v", at.fails)
	}
}

func TestLogin_UnknownUserSameErrorAndComparesDummyHash(t *testing.T) {
	h, pw, at, _ := newLogin(alice)
	_, err := h.Handle(context.Background(), Login{Identifier: "Ghost", Password: "secret1"})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
	if len(pw.compared) != 1 || pw.compared[0] != dummyHash {
		t.Fatalf("compared = %v, want one dummy comparison", pw.compared)
	}
	if at.fails[LoginSubject{Identifier: "ghost"}] != 1 {
		t.Fatalf("fails = %+v", at.fails)
	}
}

func TestLogin_UnknownUserMatchingDummyPasswordStillFails(t *testing.T) {
	h, _, _, _ := newLogin()
	if _, err := h.Handle(context.Background(), Login{Identifier: "ghost", Password: "dummy"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_BlockedEvenWithCorrectPassword(t *testing.T) {
	h, _, at, tk := newLogin(alice)
	at.blocked[LoginSubject{AccountID: 1}] = 90*time.Second + 300*time.Millisecond

	_, err := h.Handle(context.Background(), Login{Identifier: "alice@example.com", Password: "secret1"})
	var blocked *domain.TooManyAttemptsError
	if !errors.As(err, &blocked) || blocked.RetryAfterSeconds != 91 {
		t.Fatalf("err = %v, want TooManyAttemptsError{91}", err)
	}
	if len(tk.issued) != 0 {
		t.Fatal("issued tokens while blocked")
	}
}

func TestLogin_EmptyInput(t *testing.T) {
	h, _, _, _ := newLogin(alice)
	if _, err := h.Handle(context.Background(), Login{Identifier: " ", Password: "x"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("err = %v", err)
	}
}

func TestRefreshToken_ReloadsCurrentUsername(t *testing.T) {
	renamed := alice
	renamed.Username = "alice2"
	tk := &fakeTokens{parse: map[string]token.Claims{"rt": {UserID: 1, Username: "alice", Status: domain.StatusActive}}}
	h := NewRefreshTokenHandler(newFakeAccounts(renamed), tk)

	if _, err := h.Handle(context.Background(), RefreshToken{RefreshToken: "rt"}); err != nil {
		t.Fatal(err)
	}
	if tk.issued[0].Username != "alice2" {
		t.Fatalf("username = %q, want current alice2", tk.issued[0].Username)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	h := NewRefreshTokenHandler(newFakeAccounts(alice), &fakeTokens{})
	if _, err := h.Handle(context.Background(), RefreshToken{RefreshToken: "bad"}); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}

func TestRefreshToken_AccountGone(t *testing.T) {
	tk := &fakeTokens{parse: map[string]token.Claims{"rt": {UserID: 42}}}
	h := NewRefreshTokenHandler(newFakeAccounts(), tk)
	if _, err := h.Handle(context.Background(), RefreshToken{RefreshToken: "rt"}); !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}
