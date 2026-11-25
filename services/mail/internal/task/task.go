// Package task defines the mail tasks this service consumes from RabbitMQ.
// A task holds the recipient address and the OTP: never log its content.
package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

// TypeOTP is an email carrying a one-time password.
const TypeOTP = "otp"

// maxOTPLen only guards against garbage; the OTP length itself is the user
// module's policy (6 digits today).
const maxOTPLen = 12

// Task is the JSON body of one mail task.
type Task struct {
	Type      string    `json:"type"`
	To        string    `json:"to"`
	OTP       string    `json:"otp"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ErrInvalid means the task can never be processed, whatever the retry.
var ErrInvalid = errors.New("invalid task")

// Decode parses and validates a task body. Error messages never contain the
// task's content, so they are safe to log.
func Decode(body []byte) (Task, error) {
	var t Task
	if err := json.Unmarshal(body, &t); err != nil {
		// The json error can quote part of the body: do not wrap it.
		return Task{}, invalid("malformed json")
	}
	if t.Type != TypeOTP {
		return Task{}, invalid("unknown type")
	}
	if !validAddress(t.To) {
		return Task{}, invalid("bad recipient address")
	}
	if !validOTP(t.OTP) {
		return Task{}, invalid("bad otp")
	}
	if t.ExpiresAt.IsZero() {
		return Task{}, invalid("missing expires_at")
	}
	return t, nil
}

// Expired reports whether the OTP is no longer usable at now. USER-R2: an OTP
// mail not sent before the OTP expires is not sent anymore.
func (t Task) Expired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}

func invalid(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalid, reason)
}

// validAddress accepts a bare address only ("a@b.com", not "A <a@b.com>"), so
// the value can go into the To header and RCPT command as is. Rejecting CR/LF
// prevents header injection.
func validAddress(s string) bool {
	if s == "" || strings.ContainsAny(s, "\r\n") {
		return false
	}
	a, err := mail.ParseAddress(s)
	return err == nil && a.Name == "" && a.Address == s
}

func validOTP(s string) bool {
	if s == "" || len(s) > maxOTPLen {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
