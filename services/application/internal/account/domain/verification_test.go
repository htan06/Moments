package domain

import (
	"errors"
	"testing"
	"time"
)

var testPolicy = OTPPolicy{TTL: 10 * time.Minute, ResendCooldown: 60 * time.Second, MaxAttempts: 5}

var t0 = time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

func newSession(dummy bool) VerificationSession {
	return NewVerificationSession("sid", VerificationRegister, "a@x.com", "123456", dummy, t0, testPolicy)
}

func TestVerificationSession_StoresOnlyHash(t *testing.T) {
	s := newSession(false)
	if s.OTPHash == "123456" || s.OTPHash == "" {
		t.Fatalf("OTPHash = %q, want a hash", s.OTPHash)
	}
	if !s.ExpiresAt.Equal(t0.Add(10 * time.Minute)) {
		t.Errorf("ExpiresAt = %v", s.ExpiresAt)
	}
}

func TestNewRegisterSession_KeepsRegistration(t *testing.T) {
	reg := Registration{Username: "alice", Name: "Alice", PasswordHash: "hash"}
	s := NewRegisterSession("sid", "a@x.com", reg, "123456", false, t0, testPolicy)
	if s.Type != VerificationRegister || s.Email != "a@x.com" || s.Registration != reg {
		t.Fatalf("session = %+v", s)
	}
}

func TestVerify_CorrectOTPDoesNotCountAttempt(t *testing.T) {
	s := newSession(false)
	res, err := s.Verify("123456", testPolicy)
	if err != nil || res != VerifyOK || s.FailedAttempts != 0 {
		t.Fatalf("Verify = %v, %v, attempts %d", res, err, s.FailedAttempts)
	}
}

func TestVerify_WrongOTPDeletesAfterMaxAttempts(t *testing.T) {
	s := newSession(false)
	for i := 1; i < testPolicy.MaxAttempts; i++ {
		res, err := s.Verify("000000", testPolicy)
		if !errors.Is(err, ErrWrongOTP) || res != VerifyWrong {
			t.Fatalf("attempt %d: %v, %v", i, res, err)
		}
	}
	res, err := s.Verify("000000", testPolicy)
	if !errors.Is(err, ErrTooManyOTPAttempts) || res != VerifyDeleteWrong {
		t.Fatalf("last attempt: %v, %v", res, err)
	}
}

func TestVerify_DummyRejectsAnyOTP(t *testing.T) {
	s := newSession(true)
	if _, err := s.Verify("123456", testPolicy); !errors.Is(err, ErrWrongOTP) {
		t.Fatalf("dummy Verify err = %v, want ErrWrongOTP", err)
	}
}

func TestResend(t *testing.T) {
	s := newSession(false)

	var tooSoon *ResendTooSoonError
	if err := s.Resend("654321", t0.Add(20*time.Second), testPolicy); !errors.As(err, &tooSoon) || tooSoon.RetryAfterSeconds != 40 {
		t.Fatalf("early resend err = %v", err)
	}

	if err := s.Resend("654321", t0.Add(60*time.Second), testPolicy); err != nil {
		t.Fatalf("resend: %v", err)
	}
	if !s.ExpiresAt.Equal(t0.Add(10 * time.Minute)) {
		t.Errorf("resend extended the session: %v", s.ExpiresAt)
	}
	// The old otp no longer works, the new one does.
	if _, err := s.Verify("123456", testPolicy); !errors.Is(err, ErrWrongOTP) {
		t.Fatalf("old otp err = %v, want ErrWrongOTP", err)
	}
	if _, err := s.Verify("654321", testPolicy); err != nil {
		t.Fatalf("new otp err = %v", err)
	}
}

func TestExpired(t *testing.T) {
	s := newSession(false)
	if s.Expired(t0.Add(9*time.Minute)) || !s.Expired(t0.Add(10*time.Minute)) {
		t.Fatal("Expired boundary wrong")
	}
}
