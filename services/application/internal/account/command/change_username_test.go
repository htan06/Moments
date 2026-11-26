package command

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/htan06/Moments/application/internal/account/domain"
)

type renameFixture struct {
	accounts *fakeAccounts
	profiles *fakeProfiles
	tokens   *fakeTokens
	h        *ChangeUsernameHandler
}

// newRenameFixture starts with alice (id 1) and carol (id 2), each with a
// profile copy of their username.
func newRenameFixture() *renameFixture {
	carol := activeAccount(2, "carol@example.com", "carol", "secret2")
	f := &renameFixture{
		accounts: newFakeAccounts(alice, carol),
		profiles: newFakeProfiles(),
		tokens:   &fakeTokens{},
	}
	f.profiles.usernames[1] = "alice"
	f.profiles.usernames[2] = "carol"
	f.h = NewChangeUsernameHandler(f.accounts, f.profiles, f.tokens, domain.NewUsernameRules([]string{"admin"}))
	return f
}

// assertUsername checks the account and its profile copy agree on want.
func (f *renameFixture) assertUsername(t *testing.T, want string) {
	t.Helper()
	if got := f.accounts.byID[1].Username; got != want {
		t.Errorf("account username = %q, want %q", got, want)
	}
	if got := f.profiles.usernames[1]; got != want {
		t.Errorf("profile copy = %q, want %q", got, want)
	}
}

func TestChangeUsername_RenamesBothAndIssuesTokens(t *testing.T) {
	f := newRenameFixture()
	pair, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: "Bob"})
	if err != nil {
		t.Fatal(err)
	}
	f.assertUsername(t, "bob")
	if pair.AccessToken != "access-bob" || f.tokens.issued[0].Username != "bob" {
		t.Fatalf("pair = %+v, issued = %+v; want tokens with the new username", pair, f.tokens.issued)
	}
}

func TestChangeUsername_SameUsernameChangesNothing(t *testing.T) {
	f := newRenameFixture()
	if _, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: "ALICE"}); err != nil {
		t.Fatal(err)
	}
	if f.profiles.renames != 0 {
		t.Fatalf("profile renamed %d times, want 0", f.profiles.renames)
	}
	if len(f.tokens.issued) != 1 {
		t.Fatal("no token pair issued")
	}
}

func TestChangeUsername_TakenByAnotherAccount(t *testing.T) {
	f := newRenameFixture()
	if _, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: "carol"}); !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("err = %v, want ErrUsernameTaken", err)
	}
	f.assertUsername(t, "alice")
	if f.profiles.renames != 0 {
		t.Fatal("profile called although the account update failed")
	}
}

func TestChangeUsername_ProfileCopyTakenRollsBack(t *testing.T) {
	f := newRenameFixture()
	f.profiles.renameErrs = []error{domain.ErrUsernameTaken}
	if _, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: "bob"}); !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("err = %v, want ErrUsernameTaken", err)
	}
	f.assertUsername(t, "alice")
	if len(f.tokens.issued) != 0 {
		t.Fatal("tokens issued for a failed rename")
	}
}

func TestChangeUsername_ProfileErrorRollsBack(t *testing.T) {
	f := newRenameFixture()
	f.profiles.renameErrs = []error{errors.New("user module down")}
	_, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: "bob"})
	if err == nil || errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("err = %v, want an internal error", err)
	}
	f.assertUsername(t, "alice")
}

func TestChangeUsername_CommitFailureUndoesProfile(t *testing.T) {
	f := newRenameFixture()
	f.accounts.commitErr = errors.New("commit failed")
	if _, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: "bob"}); err == nil {
		t.Fatal("want error")
	}
	// The profile was renamed inside the transaction, then put back.
	if f.profiles.renames != 2 {
		t.Fatalf("profile renames = %d, want 2 (rename + undo)", f.profiles.renames)
	}
	f.assertUsername(t, "alice")
}

func TestChangeUsername_CommitAndUndoFailReportsBoth(t *testing.T) {
	f := newRenameFixture()
	f.accounts.commitErr = errors.New("commit failed")
	f.profiles.renameErrs = []error{nil, errors.New("user module down")}
	_, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: "bob"})
	if err == nil || !strings.Contains(err.Error(), "undo profile username") {
		t.Fatalf("err = %v, want it to mention the failed undo", err)
	}
}

func TestChangeUsername_InvalidUsername(t *testing.T) {
	for _, u := range []string{"ab", "admin", "a..b", ".alice"} {
		t.Run(u, func(t *testing.T) {
			f := newRenameFixture()
			if _, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: u}); !errors.Is(err, domain.ErrInvalidUsername) {
				t.Fatalf("err = %v, want ErrInvalidUsername", err)
			}
			f.assertUsername(t, "alice")
		})
	}
}

func TestChangeUsername_AccountNotActive(t *testing.T) {
	f := newRenameFixture()
	a := f.accounts.byID[1]
	a.Status = "BANNED"
	f.accounts.byID[1] = a
	if _, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 1, Username: "bob"}); !errors.Is(err, domain.ErrAccountInactive) {
		t.Fatalf("err = %v, want ErrAccountInactive", err)
	}
}

func TestChangeUsername_AccountGone(t *testing.T) {
	f := newRenameFixture()
	if _, err := f.h.Handle(context.Background(), ChangeUsername{UserID: 9, Username: "bob"}); !errors.Is(err, domain.ErrAccountNotFound) {
		t.Fatalf("err = %v, want ErrAccountNotFound", err)
	}
}
