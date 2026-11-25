package command

import (
	"context"
	"time"

	"github.com/htan06/Moments/application/internal/user/domain"
)

// VerifyRegisterOTP checks the OTP of a registration session.
type VerifyRegisterOTP struct {
	SessionID string
	OTP       string
}

type VerifyRegisterOTPHandler struct {
	sessions VerificationSessions
	policy   domain.OTPPolicy
	now      func() time.Time
}

func NewVerifyRegisterOTPHandler(sessions VerificationSessions, policy domain.OTPPolicy) *VerifyRegisterOTPHandler {
	return &VerifyRegisterOTPHandler{sessions: sessions, policy: policy, now: time.Now}
}

// Handle marks the session OTP_VERIFIED on a correct OTP. Wrong OTPs are
// counted; reaching the limit deletes the session (USER-R2).
func (h *VerifyRegisterOTPHandler) Handle(ctx context.Context, cmd VerifyRegisterOTP) error {
	return h.sessions.Update(ctx, cmd.SessionID, func(s *domain.VerificationSession) (Change, error) {
		if err := checkSession(s, domain.VerificationRegister, h.now()); err != nil {
			return KeepSession, err
		}
		res, err := s.Verify(cmd.OTP, h.policy)
		switch {
		case res == domain.VerifyDeleteWrong:
			return DeleteSession, err
		case err != nil && res == domain.VerifyOK:
			// Already verified: nothing changed.
			return KeepSession, err
		default:
			return SaveSession, err
		}
	})
}
