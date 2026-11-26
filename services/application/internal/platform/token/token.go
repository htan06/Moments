// Package token issues and verifies RS256 JWTs. Access and refresh tokens use
// separate key pairs, so one can never be accepted as the other.
// Tokens carry only user_id, username, user_status. Tokens are never logged.
package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalid = errors.New("token: invalid")
	ErrExpired = errors.New("token: expired")
)

// Claims is the identity carried by a token.
type Claims struct {
	UserID   int64
	Username string
	Status   string
}

type Pair struct {
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type Manager struct {
	keys       Keys
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

func NewManager(keys Keys, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{keys: keys, accessTTL: accessTTL, refreshTTL: refreshTTL, now: time.Now}
}

// Issue signs a new access + refresh token pair.
func (m *Manager) Issue(c Claims) (Pair, error) {
	now := m.now()
	access, accessExp, err := sign(c, m.keys.AccessPrivate, now, m.accessTTL)
	if err != nil {
		return Pair{}, err
	}
	refresh, refreshExp, err := sign(c, m.keys.RefreshPrivate, now, m.refreshTTL)
	if err != nil {
		return Pair{}, err
	}
	return Pair{
		AccessToken:      access,
		AccessExpiresAt:  accessExp,
		RefreshToken:     refresh,
		RefreshExpiresAt: refreshExp,
	}, nil
}

// ParseAccess verifies an access token.
func (m *Manager) ParseAccess(s string) (Claims, error) {
	return m.parse(s, m.keys.AccessPublic)
}

// ParseRefresh verifies a refresh token.
func (m *Manager) ParseRefresh(s string) (Claims, error) {
	return m.parse(s, m.keys.RefreshPublic)
}

type jwtClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Status   string `json:"user_status"`
	jwt.RegisteredClaims
}

func sign(c Claims, key *rsa.PrivateKey, now time.Time, ttl time.Duration) (string, time.Time, error) {
	exp := now.Add(ttl)
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims{
		UserID:   c.UserID,
		Username: c.Username,
		Status:   c.Status,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	})
	s, err := t.SignedString(key)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token: sign: %w", err)
	}
	return s, exp, nil
}

func (m *Manager) parse(s string, key *rsa.PublicKey) (Claims, error) {
	var jc jwtClaims
	_, err := jwt.ParseWithClaims(s, &jc,
		func(*jwt.Token) (any, error) { return key, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(m.now),
	)
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return Claims{}, ErrExpired
	case err != nil:
		return Claims{}, ErrInvalid
	case jc.UserID <= 0:
		return Claims{}, ErrInvalid
	}
	return Claims{UserID: jc.UserID, Username: jc.Username, Status: jc.Status}, nil
}
