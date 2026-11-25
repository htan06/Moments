package command

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
)

type registerFixture struct {
	accounts  *fakeAccounts
	profiles  *fakeProfiles
	sessions  *fakeSessions
	sender    *fakeSender
	random    *fakeRandom
	passwords *fakePasswords
	tokens    *fakeTokens

	start  *StartRegisterHandler
	resend *ResendRegisterOTPHandler
	verify *VerifyRegisterHandler
}

func newRegisterFixture(accounts ...domain.Account) *registerFixture {
	f := &registerFixture{
		accounts:  newFakeAccounts(accounts...),
		profiles:  newFakeProfiles(),
		sessions:  newFakeSessions(),
		sender:    &fakeSender{},
		random:    &fakeRandom{otp: "123456"},
		passwords: &fakePasswords{},
		tokens:    &fakeTokens{},
	}
	now := func() time.Time { return testNow }
	f.start = NewStartRegisterHandler(f.accounts, f.sessions, f.passwords, f.profiles, f.sender, syncAsync, f.random,
		domain.NewUsernameRules([]string{"admin"}), testPolicy, testLog)
	f.start.now = now
	f.resend = NewResendRegisterOTPHandler(f.sessions, f.sender, syncAsync, f.random, testPolicy, testLog)
	f.resend.now = now
	f.verify = NewVerifyRegisterHandler(f.accounts, f.profiles, f.sessions, f.tokens, testPolicy)
	f.verify.now = now
	return f
}

// validRegister is a valid submission; tests change one field at a time.
func validRegister(email string) StartRegister {
	return StartRegister{Username: "Alice", Name: " Alice Nguyễn ", Email: email, Password: "secret1"}
}

func (f *registerFixture) startSession(t *testing.T, email string) string {
	t.Helper()
	id, err := f.start.Handle(context.Background(), validRegister(email))
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestStartRegister_NewEmailKeepsDataAndSendsOTP(t *testing.T) {
	f := newRegisterFixture()
	id, err := f.start.Handle(context.Background(), validRegister(" New@Example.com "))
	if err != nil || id == "" {
		t.Fatalf("Handle = %q, %v", id, err)
	}
	if len(f.sender.sent) != 1 || f.sender.sent[0] != (sentOTP{"new@example.com", "123456"}) {
		t.Fatalf("sent = %+v", f.sender.sent)
	}
	s := f.sessions.m[id]
	want := domain.Registration{Username: "alice", Name: "Alice Nguyễn", PasswordHash: "hash:secret1"}
	if s.IsDummy || s.Email != "new@example.com" || s.Registration != want {
		t.Fatalf("session = %+v", s)
	}
}

func TestStartRegister_ExistingEmailCreatesDummySession(t *testing.T) {
	f := newRegisterFixture(activeAccount(1, "taken@example.com", "taken", "abc123"))
	id, err := f.start.Handle(context.Background(), validRegister("taken@example.com"))
	if err != nil || id == "" {
		t.Fatalf("Handle = %q, %v; want a session id like a new email", id, err)
	}
	if len(f.sender.sent) != 0 {
		t.Fatalf("dummy session sent otp: %+v", f.sender.sent)
	}
	// Even the generated otp is rejected, and no account is created.
	_, err = f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"})
	if !errors.Is(err, domain.ErrWrongOTP) {
		t.Fatalf("verify dummy err = %v, want ErrWrongOTP", err)
	}
	if len(f.accounts.byID) != 1 {
		t.Fatalf("accounts = %d, want 1", len(f.accounts.byID))
	}
}

func TestStartRegister_ExistingEmailStillHashesPassword(t *testing.T) {
	f := newRegisterFixture(activeAccount(1, "taken@example.com", "taken", "abc123"))
	if _, err := f.start.Handle(context.Background(), validRegister("taken@example.com")); err != nil {
		t.Fatal(err)
	}
	// Same bcrypt work as a new email, so timing does not reveal it (ACCOUNT-R1).
	if f.passwords.hashed != 1 {
		t.Fatalf("hashed %d times, want 1", f.passwords.hashed)
	}
}

func TestStartRegister_SendFailureDoesNotChangeResponse(t *testing.T) {
	f := newRegisterFixture()
	f.sender.err = errors.New("broker down")
	if _, err := f.start.Handle(context.Background(), validRegister("a@example.com")); err != nil {
		t.Fatalf("Handle err = %v, want nil", err)
	}
}

func TestStartRegister_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		edit func(*StartRegister)
		want error
	}{
		{"email", func(c *StartRegister) { c.Email = "nope" }, domain.ErrInvalidEmail},
		{"reserved username", func(c *StartRegister) { c.Username = "admin" }, domain.ErrInvalidUsername},
		{"blank name", func(c *StartRegister) { c.Name = "  " }, domain.ErrInvalidName},
		{"weak password", func(c *StartRegister) { c.Password = "short" }, domain.ErrInvalidPassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newRegisterFixture()
			cmd := validRegister("a@example.com")
			tt.edit(&cmd)
			if _, err := f.start.Handle(context.Background(), cmd); !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if len(f.sessions.m) != 0 || len(f.sender.sent) != 0 {
				t.Fatal("invalid input created a session or sent a mail")
			}
		})
	}
}

func TestResendRegisterOTP_TooSoon(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.resend.now = func() time.Time { return testNow.Add(10 * time.Second) }

	var tooSoon *domain.ResendTooSoonError
	err := f.resend.Handle(context.Background(), ResendRegisterOTP{SessionID: id})
	if !errors.As(err, &tooSoon) || tooSoon.RetryAfterSeconds != 50 {
		t.Fatalf("err = %v, want ResendTooSoonError{50}", err)
	}
	if len(f.sender.sent) != 1 {
		t.Fatalf("sent %d mails, want 1", len(f.sender.sent))
	}
}

func TestResendRegisterOTP_SendsNewOTP(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.resend.now = func() time.Time { return testNow.Add(time.Minute) }
	f.random.otp = "999999"

	if err := f.resend.Handle(context.Background(), ResendRegisterOTP{SessionID: id}); err != nil {
		t.Fatal(err)
	}
	if got := f.sender.sent[len(f.sender.sent)-1].otp; got != "999999" {
		t.Fatalf("resent otp = %q", got)
	}
	if _, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"}); !errors.Is(err, domain.ErrWrongOTP) {
		t.Fatalf("old otp err = %v, want ErrWrongOTP", err)
	}
	if _, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "999999"}); err != nil {
		t.Fatalf("new otp err = %v", err)
	}
}

func TestResendRegisterOTP_ExpiredSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.resend.now = func() time.Time { return testNow.Add(11 * time.Minute) }
	if err := f.resend.Handle(context.Background(), ResendRegisterOTP{SessionID: id}); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestVerifyRegister_CreatesAccount(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")

	pair, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"})
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken != "access-alice" {
		t.Errorf("pair = %+v", pair)
	}
	a, err := f.accounts.ByUsername(context.Background(), "alice")
	if err != nil || a.Email != "a@example.com" || a.PasswordHash != "hash:secret1" || a.Status != domain.StatusActive {
		t.Fatalf("account = %+v, %v", a, err)
	}
	if got := f.profiles.names[a.ID]; got != "Alice Nguyễn" {
		t.Errorf("profile name = %q", got)
	}
	if got := f.profiles.usernames[a.ID]; got != "alice" {
		t.Errorf("profile username copy = %q, want alice", got)
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Error("session not deleted after success")
	}
}

func TestVerifyRegister_WrongOTPCreatesNothing(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")

	if _, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "000000"}); !errors.Is(err, domain.ErrWrongOTP) {
		t.Fatalf("err = %v, want ErrWrongOTP", err)
	}
	if len(f.accounts.byID) != 0 {
		t.Fatal("account created with a wrong otp")
	}
	if s := f.sessions.m[id]; s.FailedAttempts != 1 {
		t.Fatalf("failed attempts = %d, want 1", s.FailedAttempts)
	}
}

func TestVerifyRegister_TooManyAttemptsDeletesSession(t *testing.T) {
	f := newRegisterFixture()
	ctx := context.Background()
	id := f.startSession(t, "a@example.com")

	for range testPolicy.MaxAttempts - 1 {
		if _, err := f.verify.Handle(ctx, VerifyRegister{SessionID: id, OTP: "000000"}); !errors.Is(err, domain.ErrWrongOTP) {
			t.Fatalf("err = %v, want ErrWrongOTP", err)
		}
	}
	if _, err := f.verify.Handle(ctx, VerifyRegister{SessionID: id, OTP: "000000"}); !errors.Is(err, domain.ErrTooManyOTPAttempts) {
		t.Fatalf("err = %v, want ErrTooManyOTPAttempts", err)
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Fatal("session not deleted")
	}
	// Even the right otp is useless now.
	if _, err := f.verify.Handle(ctx, VerifyRegister{SessionID: id, OTP: "123456"}); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestVerifyRegister_ExpiredSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.verify.now = func() time.Time { return testNow.Add(10 * time.Minute) }
	if _, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"}); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestVerifyRegister_UsernameTakenMeanwhileDeletesSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.accounts.createErr = domain.ErrUsernameTaken

	_, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"})
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("err = %v, want ErrUsernameTaken", err)
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Fatal("session not deleted; the user must register again")
	}
}

func TestVerifyRegister_EmailRaceFailsGenericallyAndDeletesSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.accounts.createErr = domain.ErrEmailTaken

	_, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"})
	if !errors.Is(err, domain.ErrRegistrationFailed) {
		t.Fatalf("err = %v, want ErrRegistrationFailed", err)
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Fatal("session not deleted")
	}
}

func TestVerifyRegister_StoreErrorKeepsSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.accounts.createErr = errors.New("db down")

	if _, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"}); err == nil {
		t.Fatal("want error")
	}
	if _, ok := f.sessions.m[id]; !ok {
		t.Fatal("session deleted on a transient error; the user should retry with the same otp")
	}
}

func TestVerifyRegister_ProfileFailureUndoesAccountAndKeepsSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.profiles.createErr = errors.New("user module down")

	if _, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"}); err == nil {
		t.Fatal("want error")
	}
	if len(f.accounts.byID) != 0 {
		t.Fatal("account kept without a profile (ACCOUNT-R7)")
	}
	if _, ok := f.sessions.m[id]; !ok {
		t.Fatal("session deleted; the user should retry with the same otp")
	}

	// Retry with the same otp once the user module is back.
	f.profiles.createErr = nil
	if _, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"}); err != nil {
		t.Fatalf("retry err = %v", err)
	}
	if len(f.accounts.byID) != 1 || len(f.profiles.names) != 1 {
		t.Fatalf("accounts = %d, profiles = %d; want 1 and 1", len(f.accounts.byID), len(f.profiles.names))
	}
}

func TestVerifyRegister_ProfileAndUndoFailReportsBoth(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.profiles.createErr = errors.New("user module down")
	f.accounts.deleteErr = errors.New("db down")

	_, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"})
	if err == nil || !strings.Contains(err.Error(), "undo account") {
		t.Fatalf("err = %v, want it to mention the failed undo", err)
	}
}

func TestVerifyRegister_ProfileUsernameTakenUndoesAccountAndDeletesSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.startSession(t, "a@example.com")
	f.profiles.createErr = domain.ErrUsernameTaken

	_, err := f.verify.Handle(context.Background(), VerifyRegister{SessionID: id, OTP: "123456"})
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("err = %v, want ErrUsernameTaken", err)
	}
	if len(f.accounts.byID) != 0 {
		t.Fatal("account kept without a profile (ACCOUNT-R7)")
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Fatal("session kept; the user must register again with another username")
	}
}
