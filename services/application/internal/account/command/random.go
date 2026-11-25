package command

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"
)

// Random generates the secrets and identifiers the use cases need. Tests
// replace it with a deterministic implementation.
type Random interface {
	SessionID() (string, error)
	OTP() (string, error)
}

// CryptoRandom uses crypto/rand.
type CryptoRandom struct{}

// SessionID returns a random UUID v4 (google/uuid, read from crypto/rand).
func (CryptoRandom) SessionID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("random session id: %w", err)
	}
	return id.String(), nil
}

// OTP returns 6 random decimal digits.
func (CryptoRandom) OTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("random otp: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
