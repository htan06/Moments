package command

import (
	"context"
	"log/slog"
	"time"

	"github.com/htan06/Moments/application/internal/account/domain"
)

// ResendOTP sends a new OTP for a pending verification session.
type ResendOTP struct {
	SessionID string
}

// ResendOTPHandler serves one flow (register or forgot password): a session of
// another flow is reported as not found.
type ResendOTPHandler struct {
	typ      domain.VerificationType
	sessions SessionUpdater
	mailer   otpMailer
	random   Random
	policy   domain.OTPPolicy
	now      func() time.Time
}

func NewResendOTPHandler(typ domain.VerificationType, sessions SessionUpdater, sender OTPSender, async Async,
	random Random, policy domain.OTPPolicy, log *slog.Logger) *ResendOTPHandler {
	return &ResendOTPHandler{
		typ:      typ,
		sessions: sessions,
		mailer:   otpMailer{sender: sender, async: async, log: log},
		random:   random,
		policy:   policy,
		now:      time.Now,
	}
}

// Handle replaces the OTP (the old one stops working) without extending the
// session. A dummy session is updated the same way but nothing is sent.
func (h *ResendOTPHandler) Handle(ctx context.Context, cmd ResendOTP) error {
	otp, err := h.random.OTP()
	if err != nil {
		return err
	}
	var sent domain.VerificationSession
	err = h.sessions.Update(ctx, cmd.SessionID, func(s *domain.VerificationSession) (Change, error) {
		if err := checkSession(s, h.typ, h.now()); err != nil {
			return KeepSession, err
		}
		if err := s.Resend(otp, h.now(), h.policy); err != nil {
			return KeepSession, err
		}
		sent = *s
		return SaveSession, nil
	})
	if err != nil {
		return err
	}
	if !sent.IsDummy {
		h.mailer.send(sent.Email, otp, sent.ExpiresAt)
	}
	return nil
}

// checkSession rejects expired sessions and sessions of another flow. Both are
// reported as not found so a session id cannot be probed across flows.
func checkSession(s *domain.VerificationSession, typ domain.VerificationType, now time.Time) error {
	if s.Expired(now) || s.Type != typ {
		return domain.ErrSessionNotFound
	}
	return nil
}
