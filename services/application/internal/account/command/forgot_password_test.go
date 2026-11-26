package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
)

type forgotFixture struct {
	accounts  *fakeAccounts
	sessions  *fakeSessions
	sender    *fakeSender
	random    *fakeRandom
	passwords *fakePasswords

	forgot *ForgotPasswordHandler
	resend *ResendOTPHandler
	reset  *ResetPasswordHandler
}

func newForgotFixture(accounts ...domain.Account) *forgotFixture {
	f := &forgotFixture{
		accounts:  newFakeAccounts(accounts...),
		sessions:  newFakeSessions(),
		sender:    &fakeSender{},
		random:    &fakeRandom{otp: "123456"},
		passwords: &fakePasswords{},
	}
	now := func() time.Time { return testNow }
	f.forgot = NewForgotPasswordHandler(f.accounts, f.sessions, f.sender, syncAsync, f.random, testPolicy, testLog)
	f.forgot.now = now
	f.resend = NewResendOTPHandler(domain.VerificationForgotPassword, f.sessions, f.sender, syncAsync, f.random,
		testPolicy, testLog)
	f.resend.now = now
	f.reset = NewResetPasswordHandler(f.accounts, f.sessions, f.passwords, testPolicy)
	f.reset.now = now
	return f
}

func (f *forgotFixture) startSession(t *testing.T, email string) string {
	t.Helper()
	id, err := f.forgot.Handle(context.Background(), ForgotPassword{Email: email})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// aliceEmail is the email of alice (login_test.go), whose password is secret1.
const aliceEmail = "alice@example.com"

func TestForgotPassword_KnownEmailSendsOTP(t *testing.T) {
	f := newForgotFixture(alice)
	id, err := f.forgot.Handle(context.Background(), ForgotPassword{Email: " Alice@Example.com "})
	if err != nil || id == "" {
		t.Fatalf("Handle = %q, %v", id, err)
	}
	if len(f.sender.sent) != 1 || f.sender.sent[0] != (sentOTP{aliceEmail, "123456"}) {
		t.Fatalf("sent = %+v", f.sender.sent)
	}
	if s := f.sessions.m[id]; s.IsDummy || s.Type != domain.VerificationForgotPassword || s.Email != aliceEmail {
		t.Fatalf("session = %+v", s)
	}
}

func TestForgotPassword_UnknownEmailCreatesDummySession(t *testing.T) {
	f := newForgotFixture(alice)
	id, err := f.forgot.Handle(context.Background(), ForgotPassword{Email: "nobody@example.com"})
	if err != nil || id == "" {
		t.Fatalf("Handle = %q, %v; want a session id like a known email", id, err)
	}
	if len(f.sender.sent) != 0 {
		t.Fatalf("dummy session sent otp: %+v", f.sender.sent)
	}
	if !f.sessions.m[id].IsDummy {
		t.Fatal("session is not a dummy")
	}
}

func TestForgotPassword_InvalidEmail(t *testing.T) {
	f := newForgotFixture()
	if _, err := f.forgot.Handle(context.Background(), ForgotPassword{Email: "nope"}); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("err = %v, want ErrInvalidEmail", err)
	}
	if len(f.sessions.m) != 0 {
		t.Fatal("invalid email created a session")
	}
}

func TestForgotPassword_SendFailureDoesNotChangeResponse(t *testing.T) {
	f := newForgotFixture(alice)
	f.sender.err = errors.New("broker down")
	if _, err := f.forgot.Handle(context.Background(), ForgotPassword{Email: aliceEmail}); err != nil {
		t.Fatalf("Handle err = %v, want nil", err)
	}
}

func TestResetPassword_SetsNewPasswordAndDeletesSession(t *testing.T) {
	f := newForgotFixture(alice)
	id := f.startSession(t, aliceEmail)

	if err := f.reset.Handle(context.Background(), ResetPassword{SessionID: id, OTP: "123456", NewPassword: "new1pass"}); err != nil {
		t.Fatal(err)
	}
	if got := f.accounts.byID[1].PasswordHash; got != "hash:new1pass" {
		t.Fatalf("password hash = %q", got)
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Fatal("session not deleted after success")
	}
}

func TestResetPassword_SamePasswordAsBeforeIsAllowed(t *testing.T) {
	f := newForgotFixture(alice)
	id := f.startSession(t, aliceEmail)
	if err := f.reset.Handle(context.Background(), ResetPassword{SessionID: id, OTP: "123456", NewPassword: "secret1"}); err != nil {
		t.Fatalf("err = %v; reset must not reveal that it matches the old password", err)
	}
}

func TestResetPassword_WrongOTPChangesNothing(t *testing.T) {
	f := newForgotFixture(alice)
	id := f.startSession(t, aliceEmail)

	err := f.reset.Handle(context.Background(), ResetPassword{SessionID: id, OTP: "000000", NewPassword: "new1pass"})
	if !errors.Is(err, domain.ErrWrongOTP) {
		t.Fatalf("err = %v, want ErrWrongOTP", err)
	}
	if got := f.accounts.byID[1].PasswordHash; got != "hash:secret1" {
		t.Fatalf("password changed with a wrong otp: %q", got)
	}
	if s := f.sessions.m[id]; s.FailedAttempts != 1 {
		t.Fatalf("failed attempts = %d, want 1", s.FailedAttempts)
	}
}

func TestResetPassword_DummySessionRejectsEveryOTP(t *testing.T) {
	f := newForgotFixture(alice)
	id := f.startSession(t, "nobody@example.com")
	err := f.reset.Handle(context.Background(), ResetPassword{SessionID: id, OTP: "123456", NewPassword: "new1pass"})
	if !errors.Is(err, domain.ErrWrongOTP) {
		t.Fatalf("err = %v, want ErrWrongOTP like a real wrong otp", err)
	}
}

func TestResetPassword_TooManyAttemptsDeletesSession(t *testing.T) {
	f := newForgotFixture(alice)
	ctx := context.Background()
	id := f.startSession(t, aliceEmail)
	wrong := ResetPassword{SessionID: id, OTP: "000000", NewPassword: "new1pass"}

	for range testPolicy.MaxAttempts - 1 {
		if err := f.reset.Handle(ctx, wrong); !errors.Is(err, domain.ErrWrongOTP) {
			t.Fatalf("err = %v, want ErrWrongOTP", err)
		}
	}
	if err := f.reset.Handle(ctx, wrong); !errors.Is(err, domain.ErrTooManyOTPAttempts) {
		t.Fatalf("err = %v, want ErrTooManyOTPAttempts", err)
	}
	err := f.reset.Handle(ctx, ResetPassword{SessionID: id, OTP: "123456", NewPassword: "new1pass"})
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestResetPassword_InvalidPasswordDoesNotUseAnAttempt(t *testing.T) {
	f := newForgotFixture(alice)
	id := f.startSession(t, aliceEmail)

	err := f.reset.Handle(context.Background(), ResetPassword{SessionID: id, OTP: "000000", NewPassword: "short"})
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Fatalf("err = %v, want ErrInvalidPassword", err)
	}
	if s := f.sessions.m[id]; s.FailedAttempts != 0 {
		t.Fatalf("failed attempts = %d, want 0", s.FailedAttempts)
	}
}

func TestResetPassword_RegisterSessionIsNotFound(t *testing.T) {
	f := newForgotFixture(alice)
	reg := domain.NewRegisterSession("reg1", "new@example.com", domain.Registration{}, "123456", false, testNow, testPolicy)
	f.sessions.m[reg.ID] = reg

	err := f.reset.Handle(context.Background(), ResetPassword{SessionID: reg.ID, OTP: "123456", NewPassword: "new1pass"})
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestResetPassword_ExpiredSession(t *testing.T) {
	f := newForgotFixture(alice)
	id := f.startSession(t, aliceEmail)
	f.reset.now = func() time.Time { return testNow.Add(10 * time.Minute) }
	err := f.reset.Handle(context.Background(), ResetPassword{SessionID: id, OTP: "123456", NewPassword: "new1pass"})
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestResetPassword_StoreErrorKeepsSession(t *testing.T) {
	f := newForgotFixture(alice)
	id := f.startSession(t, aliceEmail)
	f.accounts.updateErr = errors.New("db down")

	if err := f.reset.Handle(context.Background(), ResetPassword{SessionID: id, OTP: "123456", NewPassword: "new1pass"}); err == nil {
		t.Fatal("want error")
	}
	if _, ok := f.sessions.m[id]; !ok {
		t.Fatal("session deleted on a transient error; the user should retry with the same otp")
	}
}

func TestResetPassword_AccountGoneDeletesSession(t *testing.T) {
	f := newForgotFixture(alice)
	id := f.startSession(t, aliceEmail)
	delete(f.accounts.byID, 1)

	err := f.reset.Handle(context.Background(), ResetPassword{SessionID: id, OTP: "123456", NewPassword: "new1pass"})
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
	if _, ok := f.sessions.m[id]; ok {
		t.Fatal("session kept for a missing account")
	}
}

func TestResendForgotOTP_RegisterSessionIsNotFound(t *testing.T) {
	f := newForgotFixture()
	reg := domain.NewRegisterSession("reg1", "new@example.com", domain.Registration{}, "123456", false, testNow, testPolicy)
	f.sessions.m[reg.ID] = reg
	f.resend.now = func() time.Time { return testNow.Add(time.Minute) }

	if err := f.resend.Handle(context.Background(), ResendOTP{SessionID: reg.ID}); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestResendForgotOTP_NewOTPReplacesOld(t *testing.T) {
	f := newForgotFixture(alice)
	ctx := context.Background()
	id := f.startSession(t, aliceEmail)
	f.resend.now = func() time.Time { return testNow.Add(time.Minute) }
	f.random.otp = "999999"

	if err := f.resend.Handle(ctx, ResendOTP{SessionID: id}); err != nil {
		t.Fatal(err)
	}
	if err := f.reset.Handle(ctx, ResetPassword{SessionID: id, OTP: "123456", NewPassword: "new1pass"}); !errors.Is(err, domain.ErrWrongOTP) {
		t.Fatalf("old otp err = %v, want ErrWrongOTP", err)
	}
	if err := f.reset.Handle(ctx, ResetPassword{SessionID: id, OTP: "999999", NewPassword: "new1pass"}); err != nil {
		t.Fatalf("new otp err = %v", err)
	}
}
