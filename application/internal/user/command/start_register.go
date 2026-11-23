package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/htan06/Moments/application/internal/user/domain"
)

// StartRegister is step 1 of registration: send an OTP to the email.
type StartRegister struct {
	Email string
}

type StartRegisterHandler struct {
	accounts Accounts
	sessions VerificationSessions
	mailer   otpMailer
	random   Random
	policy   domain.OTPPolicy
	now      func() time.Time
}

func NewStartRegisterHandler(accounts Accounts, sessions VerificationSessions, sender OTPSender, async Async,
	random Random, policy domain.OTPPolicy, log *slog.Logger) *StartRegisterHandler {
	return &StartRegisterHandler{
		accounts: accounts,
		sessions: sessions,
		mailer:   otpMailer{sender: sender, async: async, log: log},
		random:   random,
		policy:   policy,
		now:      time.Now,
	}
}

// Handle always returns a session id for a well-formed email. If the email is
// already registered the session is a dummy: no OTP is sent and every OTP is
// wrong (USER-R1).
func (h *StartRegisterHandler) Handle(ctx context.Context, cmd StartRegister) (sessionID string, err error) {
	email, err := domain.ParseEmail(cmd.Email)
	if err != nil {
		return "", err
	}
	exists, err := h.accounts.EmailExists(ctx, email)
	if err != nil {
		return "", fmt.Errorf("start register: %w", err)
	}
	id, err := h.random.SessionID()
	if err != nil {
		return "", err
	}
	// An otp is generated for dummy sessions too so both branches do the same work.
	otp, err := h.random.OTP()
	if err != nil {
		return "", err
	}
	s := domain.NewVerificationSession(id, domain.VerificationRegister, email, otp, exists, h.now(), h.policy)
	if err := h.sessions.Create(ctx, s); err != nil {
		return "", fmt.Errorf("start register: %w", err)
	}
	if !s.IsDummy {
		h.mailer.send(email, otp, s.ExpiresAt)
	}
	return s.ID, nil
}
