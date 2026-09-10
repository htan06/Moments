package security

import "math/rand/v2"

const OTPCharset = "0123456789"

type OTPProvider struct {
	charset       string
	charsetLength int
}

func NewOTPProvider() *OTPProvider {
	return &OTPProvider{
		charset:       OTPCharset,
		charsetLength: len(OTPCharset),
	}
}

func (sr *OTPProvider) RandOTP() string {
	otp := make([]byte, 6)
	for i := 0; i < 6; i++ {
		randIndex := rand.IntN(sr.charsetLength)
		otp[i] = sr.charset[randIndex]
	}
	return string(otp)
}
