package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// ForgotPassword starts a password reset: an otp is sent to the email and
// checked by ResetPassword together with the new password.
type ForgotPassword struct {
	Email string
}

type ForgotPasswordHandler struct {
	accounts AccountReader
	sessions SessionCreator
	mailer   otpMailer
	random   Random
	policy   domain.OTPPolicy
	now      func() time.Time
}

func NewForgotPasswordHandler(accounts AccountReader, sessions SessionCreator, sender OTPSender, async Async,
	random Random, policy domain.OTPPolicy, log *slog.Logger) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{
		accounts: accounts,
		sessions: sessions,
		mailer:   otpMailer{sender: sender, async: async, log: log},
		random:   random,
		policy:   policy,
		now:      time.Now,
	}
}

// Handle always returns a session id for a well-formed email. If no account
// has this email the session is a dummy: no otp is sent and every otp is wrong
// (ACCOUNT-R1). Both branches do the same work (one lookup, one session write)
// and the mail is sent in the background, so timing does not tell them apart.
func (h *ForgotPasswordHandler) Handle(ctx context.Context, cmd ForgotPassword) (sessionID string, err error) {
	email, err := domain.ParseEmail(cmd.Email)
	if err != nil {
		return "", err
	}
	exists, err := h.accounts.EmailExists(ctx, email)
	if err != nil {
		return "", fmt.Errorf("forgot password: %w", err)
	}
	id, err := h.random.SessionID()
	if err != nil {
		return "", err
	}
	otp, err := h.random.OTP()
	if err != nil {
		return "", err
	}
	s := domain.NewVerificationSession(id, domain.VerificationForgotPassword, email, otp, !exists, h.now(), h.policy)
	if err := h.sessions.Create(ctx, s); err != nil {
		return "", fmt.Errorf("forgot password: %w", err)
	}
	if !s.IsDummy {
		h.mailer.send(email, otp, s.ExpiresAt)
	}
	return s.ID, nil
}
