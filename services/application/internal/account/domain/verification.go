package domain

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"time"
)

// VerificationType is the flow a verification session belongs to.
type VerificationType string

const (
	VerificationRegister       VerificationType = "REGISTER"
	VerificationForgotPassword VerificationType = "FORGOT_PASSWORD"
)

// OTPPolicy holds the configurable OTP rules (ACCOUNT-R2).
type OTPPolicy struct {
	TTL            time.Duration // session lifetime from creation, never extended
	ResendCooldown time.Duration // minimum gap between two sends
	MaxAttempts    int           // wrong attempts that delete the session
}

// VerificationSession is one OTP verification (register / forgot password).
// It has no status: the otp is checked in the request that finishes the flow
// (ACCOUNT-R2), and the session is deleted when the flow ends, when the otp is
// wrong too often, or when it expires.
type VerificationSession struct {
	ID    string
	Type  VerificationType
	Email string
	// Registration holds what the user submitted to register; empty for other
	// flows.
	Registration   Registration
	OTPHash        string
	FailedAttempts int
	LastSentAt     time.Time
	IsDummy        bool
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

// NewVerificationSession starts a session with a freshly generated otp.
// A dummy session (email not valid for the flow, ACCOUNT-R1) never accepts any otp.
func NewVerificationSession(id string, typ VerificationType, email, otp string, dummy bool, now time.Time, p OTPPolicy) VerificationSession {
	return VerificationSession{
		ID:         id,
		Type:       typ,
		Email:      email,
		OTPHash:    HashOTP(id, otp),
		LastSentAt: now,
		IsDummy:    dummy,
		CreatedAt:  now,
		ExpiresAt:  now.Add(p.TTL),
	}
}

// Registration is the account data submitted at the start of registration,
// already validated; the password is only kept hashed.
type Registration struct {
	Username     string
	Name         string
	PasswordHash string
}

// NewRegisterSession starts a registration session holding the submitted data
// until the otp is verified. Dummy sessions (email already registered) hold it
// too, so both branches look and cost the same (ACCOUNT-R1).
func NewRegisterSession(id, email string, reg Registration, otp string, dummy bool, now time.Time, p OTPPolicy) VerificationSession {
	s := NewVerificationSession(id, VerificationRegister, email, otp, dummy, now, p)
	s.Registration = reg
	return s
}

// Expired reports whether the session is past its lifetime.
func (s *VerificationSession) Expired(now time.Time) bool {
	return !now.Before(s.ExpiresAt)
}

// Resend replaces the otp with a new one; the old one stops working. The
// session lifetime is not extended.
func (s *VerificationSession) Resend(otp string, now time.Time, p OTPPolicy) error {
	if wait := s.LastSentAt.Add(p.ResendCooldown).Sub(now); wait > 0 {
		return &ResendTooSoonError{RetryAfterSeconds: ceilSeconds(wait)}
	}
	s.OTPHash = HashOTP(s.ID, otp)
	s.LastSentAt = now
	return nil
}

// VerifyResult tells the caller what to do with the session after Verify.
type VerifyResult int

const (
	VerifyOK          VerifyResult = iota // otp correct, the caller finishes the flow
	VerifyWrong                           // otp wrong, save the session
	VerifyDeleteWrong                     // otp wrong and attempts exhausted, delete the session
)

// Verify checks otp and counts wrong attempts. A correct otp changes nothing:
// the caller finishes the flow and deletes the session. A dummy session treats
// every otp as wrong.
func (s *VerificationSession) Verify(otp string, p OTPPolicy) (VerifyResult, error) {
	if !s.IsDummy && otpEqual(s.OTPHash, HashOTP(s.ID, otp)) {
		return VerifyOK, nil
	}
	s.FailedAttempts++
	if s.FailedAttempts >= p.MaxAttempts {
		return VerifyDeleteWrong, ErrTooManyOTPAttempts
	}
	return VerifyWrong, ErrWrongOTP
}

// HashOTP binds the otp to its session, so a hash is useless for any other
// session. Only this hash is stored, never the otp.
func HashOTP(sessionID, otp string) string {
	h := sha256.Sum256([]byte(sessionID + ":" + otp))
	return hex.EncodeToString(h[:])
}

func otpEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func ceilSeconds(d time.Duration) int {
	s := int(d / time.Second)
	if d%time.Second != 0 {
		s++
	}
	return s
}
