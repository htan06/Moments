// Package password hashes passwords with bcrypt.
package password

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Bcrypt pre-hashes with SHA-256 (base64, 44 bytes) because bcrypt only reads
// the first 72 bytes and a 64-character password can exceed that in UTF-8.
type Bcrypt struct {
	cost int
}

func NewBcrypt(cost int) (*Bcrypt, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("password: bcrypt cost %d out of range", cost)
	}
	return &Bcrypt{cost: cost}, nil
}

func (b *Bcrypt) Hash(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword(prehash(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("password: hash: %w", err)
	}
	return string(h), nil
}

func (b *Bcrypt) Matches(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), prehash(password)) == nil
}

func prehash(p string) []byte {
	sum := sha256.Sum256([]byte(p))
	out := make([]byte, base64.StdEncoding.EncodedLen(len(sum)))
	base64.StdEncoding.Encode(out, sum[:])
	return out
}
