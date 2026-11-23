package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/htan06/Moments/application/internal/user/domain"
)

type registerFixture struct {
	accounts  *fakeAccounts
	sessions  *fakeSessions
	sender    *fakeSender
	random    *fakeRandom
	passwords *fakePasswords
	tokens    *fakeTokens

	start    *StartRegisterHandler
	resend   *ResendRegisterOTPHandler
	verify   *VerifyRegisterOTPHandler
	complete *CompleteRegisterHandler
}

func newRegisterFixture(accounts ...domain.Account) *registerFixture {
	f := &registerFixture{
		accounts:  newFakeAccounts(accounts...),
		sessions:  newFakeSessions(),
		sender:    &fakeSender{},
		random:    &fakeRandom{otp: "123456"},
		passwords: &fakePasswords{},
		tokens:    &fakeTokens{},
	}
	now := func() time.Time { return testNow }
	f.start = NewStartRegisterHandler(f.accounts, f.sessions, f.sender, syncAsync, f.random, testPolicy, testLog)
	f.start.now = now
	f.resend = NewResendRegisterOTPHandler(f.sessions, f.sender, syncAsync, f.random, testPolicy, testLog)
	f.resend.now = now
	f.verify = NewVerifyRegisterOTPHandler(f.sessions, testPolicy)
	f.verify.now = now
	f.complete = NewCompleteRegisterHandler(f.accounts, f.sessions, f.passwords, f.tokens, f.random,
		domain.NewUsernameRules([]string{"admin"}))
	f.complete.now = now
	return f
}

func (f *registerFixture) verifiedSession(t *testing.T, email string) string {
	t.Helper()
	ctx := context.Background()
	id, err := f.start.Handle(ctx, StartRegister{Email: email})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.verify.Handle(ctx, VerifyRegisterOTP{SessionID: id, OTP: "123456"}); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestStartRegister_NewEmailSendsOTP(t *testing.T) {
	f := newRegisterFixture()
	id, err := f.start.Handle(context.Background(), StartRegister{Email: " New@Example.com "})
	if err != nil || id == "" {
		t.Fatalf("Handle = %q, %v", id, err)
	}
	if len(f.sender.sent) != 1 || f.sender.sent[0] != (sentOTP{"new@example.com", "123456"}) {
		t.Fatalf("sent = %+v", f.sender.sent)
	}
	if s := f.sessions.m[id]; s.IsDummy || s.Email != "new@example.com" {
		t.Fatalf("session = %+v", s)
	}
}

func TestStartRegister_ExistingEmailCreatesDummySession(t *testing.T) {
	f := newRegisterFixture(activeAccount(1, "taken@example.com", "taken", "abc123"))
	id, err := f.start.Handle(context.Background(), StartRegister{Email: "taken@example.com"})
	if err != nil || id == "" {
		t.Fatalf("Handle = %q, %v; want a session id like a new email", id, err)
	}
	if len(f.sender.sent) != 0 {
		t.Fatalf("dummy session sent otp: %+v", f.sender.sent)
	}
	// Even the generated otp is rejected.
	err = f.verify.Handle(context.Background(), VerifyRegisterOTP{SessionID: id, OTP: "123456"})
	if !errors.Is(err, domain.ErrWrongOTP) {
		t.Fatalf("verify dummy err = %v, want ErrWrongOTP", err)
	}
}

func TestStartRegister_SendFailureDoesNotChangeResponse(t *testing.T) {
	f := newRegisterFixture()
	f.sender.err = errors.New("broker down")
	if _, err := f.start.Handle(context.Background(), StartRegister{Email: "a@example.com"}); err != nil {
		t.Fatalf("Handle err = %v, want nil", err)
	}
}

func TestStartRegister_InvalidEmail(t *testing.T) {
	f := newRegisterFixture()
	if _, err := f.start.Handle(context.Background(), StartRegister{Email: "nope"}); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("err = %v, want ErrInvalidEmail", err)
	}
}

func TestResendRegisterOTP_TooSoon(t *testing.T) {
	f := newRegisterFixture()
	id, _ := f.start.Handle(context.Background(), StartRegister{Email: "a@example.com"})
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
	id, _ := f.start.Handle(context.Background(), StartRegister{Email: "a@example.com"})
	f.resend.now = func() time.Time { return testNow.Add(time.Minute) }
	f.random.otp = "999999"

	if err := f.resend.Handle(context.Background(), ResendRegisterOTP{SessionID: id}); err != nil {
		t.Fatal(err)
	}
	if got := f.sender.sent[len(f.sender.sent)-1].otp; got != "999999" {
		t.Fatalf("resent otp = %q", got)
	}
	if err := f.verify.Handle(context.Background(), VerifyRegisterOTP{SessionID: id, OTP: "123456"}); !errors.Is(err, domain.ErrWrongOTP) {
		t.Fatalf("old otp err = %v, want ErrWrongOTP", err)
	}
}

func TestResendRegisterOTP_ExpiredSession(t *testing.T) {
	f := newRegisterFixture()
	id, _ := f.start.Handle(context.Background(), StartRegister{Email: "a@example.com"})
	f.resend.now = func() time.Time { return testNow.Add(11 * time.Minute) }
	if err := f.resend.Handle(context.Background(), ResendRegisterOTP{SessionID: id}); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestVerifyRegisterOTP_TooManyAttemptsDeletesSession(t *testing.T) {
	f := newRegisterFixture()
	ctx := context.Background()
	id, _ := f.start.Handle(ctx, StartRegister{Email: "a@example.com"})

	for range testPolicy.MaxAttempts - 1 {
		if err := f.verify.Handle(ctx, VerifyRegisterOTP{SessionID: id, OTP: "000000"}); !errors.Is(err, domain.ErrWrongOTP) {
			t.Fatalf("err = %v, want ErrWrongOTP", err)
		}
	}
	if err := f.verify.Handle(ctx, VerifyRegisterOTP{SessionID: id, OTP: "000000"}); !errors.Is(err, domain.ErrTooManyOTPAttempts) {
		t.Fatalf("err = %v, want ErrTooManyOTPAttempts", err)
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Fatal("session not deleted")
	}
	// Even the right otp is useless now.
	if err := f.verify.Handle(ctx, VerifyRegisterOTP{SessionID: id, OTP: "123456"}); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestCompleteRegister_Success(t *testing.T) {
	f := newRegisterFixture()
	id := f.verifiedSession(t, "a@example.com")

	pair, err := f.complete.Handle(context.Background(), CompleteRegister{SessionID: id, Username: "Alice", Password: "secret1"})
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
	if _, ok := f.sessions.m[id]; ok {
		t.Error("session not deleted after success")
	}
}

func TestCompleteRegister_SessionNotVerified(t *testing.T) {
	f := newRegisterFixture()
	id, _ := f.start.Handle(context.Background(), StartRegister{Email: "a@example.com"})
	_, err := f.complete.Handle(context.Background(), CompleteRegister{SessionID: id, Username: "alice", Password: "secret1"})
	if !errors.Is(err, domain.ErrSessionNotVerified) {
		t.Fatalf("err = %v, want ErrSessionNotVerified", err)
	}
}

func TestCompleteRegister_InvalidInputKeepsSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.verifiedSession(t, "a@example.com")
	ctx := context.Background()

	if _, err := f.complete.Handle(ctx, CompleteRegister{SessionID: id, Username: "admin", Password: "secret1"}); !errors.Is(err, domain.ErrInvalidUsername) {
		t.Fatalf("reserved username err = %v", err)
	}
	if _, err := f.complete.Handle(ctx, CompleteRegister{SessionID: id, Username: "alice", Password: "short"}); !errors.Is(err, domain.ErrInvalidPassword) {
		t.Fatalf("weak password err = %v", err)
	}
	if _, ok := f.sessions.m[id]; !ok {
		t.Fatal("session deleted on input error")
	}
}

func TestCompleteRegister_UsernameTakenKeepsSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.verifiedSession(t, "a@example.com")
	f.accounts.createErr = domain.ErrUsernameTaken

	_, err := f.complete.Handle(context.Background(), CompleteRegister{SessionID: id, Username: "alice", Password: "secret1"})
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("err = %v, want ErrUsernameTaken", err)
	}
	if _, ok := f.sessions.m[id]; !ok {
		t.Fatal("session deleted; user should be able to pick another username")
	}
}

func TestCompleteRegister_EmailRaceFailsGenericallyAndDeletesSession(t *testing.T) {
	f := newRegisterFixture()
	id := f.verifiedSession(t, "a@example.com")
	f.accounts.createErr = domain.ErrEmailTaken

	_, err := f.complete.Handle(context.Background(), CompleteRegister{SessionID: id, Username: "alice", Password: "secret1"})
	if !errors.Is(err, domain.ErrRegistrationFailed) {
		t.Fatalf("err = %v, want ErrRegistrationFailed", err)
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Fatal("session not deleted")
	}
}
