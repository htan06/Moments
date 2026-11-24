package command

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

// Random generates the secrets and identifiers the use cases need. Tests
// replace it with a deterministic implementation.
type Random interface {
	SessionID() (string, error)
	OTP() (string, error)
	NameSuffix() (string, error)
}

// CryptoRandom uses crypto/rand.
type CryptoRandom struct{}

// SessionID returns 128 random bits, base64url encoded.
func (CryptoRandom) SessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random session id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// OTP returns 6 random decimal digits.
func (CryptoRandom) OTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("random otp: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

const nameSuffixAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// NameSuffix returns 4 random [a-z0-9] characters.
func (CryptoRandom) NameSuffix() (string, error) {
	b := make([]byte, 4)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(nameSuffixAlphabet))))
		if err != nil {
			return "", fmt.Errorf("random name suffix: %w", err)
		}
		b[i] = nameSuffixAlphabet[n.Int64()]
	}
	return string(b), nil
}
