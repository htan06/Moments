package command

import (
	"context"
	"log/slog"
	"time"
)

// Async runs fn without making the request wait for it. The request context is
// not passed on: fn gets its own.
type Async func(fn func(ctx context.Context))

// GoAsync runs fn in a goroutine with a context bounded by timeout.
func GoAsync(timeout time.Duration) Async {
	return func(fn func(ctx context.Context)) {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			fn(ctx)
		}()
	}
}

// otpMailer sends OTP emails in the background. Failures are only logged and
// never change the response (USER-R1, USER-R2); the user can ask to resend.
type otpMailer struct {
	sender OTPSender
	async  Async
	log    *slog.Logger
}

func (m otpMailer) send(email, otp string, expiresAt time.Time) {
	m.async(func(ctx context.Context) {
		if err := m.sender.SendOTP(ctx, email, otp, expiresAt); err != nil {
			// Never log the otp or the email address.
			m.log.ErrorContext(ctx, "send otp mail failed", slog.Any("err", err))
		}
	})
}
