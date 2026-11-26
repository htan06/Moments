// Package password hashes passwords with bcrypt.
package password

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Bcrypt pre-hashes with SHA-256 (base64, 44 bytes) because bcrypt only reads
// the first 72 bytes and a 64-character password can exceed that in UTF-8.
type Bcrypt struct {
	cost int
	// dummyHash is a hash of a random password, made with the real cost at
	// startup. CompareDummy compares against it so a login for an unknown user
	// costs as much as a real one (ACCOUNT-R1).
	dummyHash []byte
}

func NewBcrypt(cost int) (*Bcrypt, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("password: bcrypt cost %d out of range", cost)
	}
	dummy, err := bcrypt.GenerateFromPassword(prehash(rand.Text()), cost)
	if err != nil {
		return nil, fmt.Errorf("password: dummy hash: %w", err)
	}
	return &Bcrypt{cost: cost, dummyHash: dummy}, nil
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

// CompareDummy does the work of Matches against the dummy hash and discards
// the result. Use it when there is no real hash to compare with.
func (b *Bcrypt) CompareDummy(password string) {
	_ = bcrypt.CompareHashAndPassword(b.dummyHash, prehash(password))
}

func prehash(p string) []byte {
	sum := sha256.Sum256([]byte(p))
	out := make([]byte, base64.StdEncoding.EncodedLen(len(sum)))
	base64.StdEncoding.Encode(out, sum[:])
	return out
}
