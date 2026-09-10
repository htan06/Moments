package infra

import (
	"context"
	"fmt"

	"github.com/htan06/Moments/auth/config"
	"gopkg.in/gomail.v2"
)

type GmailOTPSender struct {
	dialer      *gomail.Dialer
	mailAddress *config.MailAddress
}

func NewGmailOTPSender(dialer *gomail.Dialer, mailAddress *config.MailAddress) *GmailOTPSender {
	return &GmailOTPSender{
		dialer:      dialer,
		mailAddress: mailAddress,
	}
}

func (gs *GmailOTPSender) SendOTP(ctx context.Context, email string, otp string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", gs.mailAddress.Email, gs.mailAddress.Name)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "OTP code")
	m.SetBody("text/plain", "Your OTP code is "+otp)

	if err := gs.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("GmailOTPSender.Send: %w", err)
	}
	return nil
}
