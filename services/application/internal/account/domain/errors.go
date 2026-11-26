// Package domain holds the account module's entities, invariants and domain
// errors. It imports only the standard library.
package domain

import "errors"

var (
	// Input validation.
	ErrInvalidEmail    = errors.New("account: invalid email")
	ErrInvalidUsername = errors.New("account: invalid username")
	ErrInvalidPassword = errors.New("account: invalid password")
	// ErrInvalidName: the profile name is rejected by the user module.
	ErrInvalidName = errors.New("account: invalid name")

	// ErrUsernameTaken: username is public, so a clash is reported precisely (ACCOUNT-R5).
	ErrUsernameTaken = errors.New("account: username taken")

	// ErrRegistrationFailed is the generic error when registration cannot
	// complete without revealing why (ACCOUNT-R1, ACCOUNT-R7: email raced in).
	ErrRegistrationFailed = errors.New("account: registration failed")

	// Change password (signed in, ACCOUNT-R4).
	ErrWrongPassword = errors.New("account: wrong current password")
	ErrSamePassword  = errors.New("account: new password equals current password")

	// ErrInvalidCredentials covers unknown identifier and wrong password (ACCOUNT-R1).
	ErrInvalidCredentials = errors.New("account: invalid credentials")

	// Verification session (ACCOUNT-R2).
	ErrSessionNotFound    = errors.New("account: verification session not found")
	ErrWrongOTP           = errors.New("account: wrong otp")
	ErrTooManyOTPAttempts = errors.New("account: too many otp attempts")
	ErrWrongSessionType   = errors.New("account: wrong verification session type")

	// ErrAccountInactive: account exists but is not ACTIVE (ACCOUNT-R6).
	ErrAccountInactive = errors.New("account: account not active")

	// Repository outcomes, never returned to clients as-is.
	ErrAccountNotFound = errors.New("account: account not found")
	ErrEmailTaken      = errors.New("account: email taken")

	// ErrInvalidToken: refresh token invalid / expired, or its user is gone.
	ErrInvalidToken = errors.New("account: invalid token")
)

// ResendTooSoonError means the resend cooldown has not elapsed (ACCOUNT-R2).
type ResendTooSoonError struct {
	RetryAfterSeconds int
}

func (e *ResendTooSoonError) Error() string { return "account: otp resend too soon" }

// TooManyAttemptsError means login (or old-password check) is blocked until
// the window ends (ACCOUNT-R3).
type TooManyAttemptsError struct {
	RetryAfterSeconds int
}

func (e *TooManyAttemptsError) Error() string { return "account: too many attempts" }
