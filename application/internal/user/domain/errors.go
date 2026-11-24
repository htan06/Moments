// Package domain holds the user module's entities, invariants and domain
// errors. It imports only the standard library.
package domain

import "errors"

var (
	// Input validation.
	ErrInvalidEmail    = errors.New("user: invalid email")
	ErrInvalidUsername = errors.New("user: invalid username")
	ErrInvalidPassword = errors.New("user: invalid password")
	ErrInvalidName     = errors.New("user: invalid name")

	// ErrUsernameTaken: username is public, so a clash is reported precisely (USER-R5).
	ErrUsernameTaken = errors.New("user: username taken")

	// ErrRegistrationFailed is the generic error when registration cannot
	// complete without revealing why (USER-R1, USER-R7: email raced in).
	ErrRegistrationFailed = errors.New("user: registration failed")

	// ErrInvalidCredentials covers unknown identifier and wrong password (USER-R1).
	ErrInvalidCredentials = errors.New("user: invalid credentials")

	// Verification session (USER-R2).
	ErrSessionNotFound    = errors.New("user: verification session not found")
	ErrWrongOTP           = errors.New("user: wrong otp")
	ErrTooManyOTPAttempts = errors.New("user: too many otp attempts")
	ErrSessionNotVerified = errors.New("user: verification session not verified")
	ErrSessionVerified    = errors.New("user: verification session already verified")
	ErrWrongSessionType   = errors.New("user: wrong verification session type")

	// ErrAccountInactive: account exists but is not ACTIVE (USER-R6).
	ErrAccountInactive = errors.New("user: account not active")

	// Repository outcomes, never returned to clients as-is.
	ErrAccountNotFound = errors.New("user: account not found")
	ErrEmailTaken      = errors.New("user: email taken")

	// ErrInvalidToken: refresh token invalid / expired, or its user is gone.
	ErrInvalidToken = errors.New("user: invalid token")
)

// ResendTooSoonError means the resend cooldown has not elapsed (USER-R2).
type ResendTooSoonError struct {
	RetryAfterSeconds int
}

func (e *ResendTooSoonError) Error() string { return "user: otp resend too soon" }

// TooManyAttemptsError means login (or old-password check) is blocked until
// the window ends (USER-R3).
type TooManyAttemptsError struct {
	RetryAfterSeconds int
}

func (e *TooManyAttemptsError) Error() string { return "user: too many attempts" }
