package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/errs"
	"github.com/htan06/Moments/internal/module/auth/domain"
)

type UserClaimsAccess struct {
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type UserClaimsRefresh struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTProvier struct {
	cfg *config.JWTConfig
}

func NewJWTProvider(cfg *config.JWTConfig) *JWTProvier {
	return &JWTProvier{
		cfg: cfg,
	}
}

func (jp *JWTProvier) GenerateAccessToken(user domain.User) (string, error) {
	claim := UserClaimsAccess{
		UserID:   user.ID,
		Name:     user.Name,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			Issuer:    "echo-authenticator",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jp.cfg.TtlAccess())),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claim)
	tokenSigned, err := token.SignedString(jp.cfg.PrivateKeyAccess())
	if err != nil {
		return "", fmt.Errorf("JWTProvider[GenerateAccessToken]: %w", err)
	}
	return tokenSigned, nil
}

func (jp *JWTProvier) ParseAccessToken(tokenString string) (UserClaimsAccess, error) {
	var claim UserClaimsAccess
	token, err := jwt.ParseWithClaims(tokenString, &claim, func(token *jwt.Token) (any, error) {
		return jp.cfg.PublicKeyAccess(), nil
	})

	if err != nil {
		return UserClaimsAccess{}, fmt.Errorf("JWTProvider[GenerateAccessToken]: %w", err)
	}

	if !token.Valid {
		return UserClaimsAccess{}, errs.NewError(errs.AuthenticationFailure, nil, domain.TokenInvalid)
	}
	return claim, nil
}

func (jp *JWTProvier) GenerateRefreshToken(user domain.User) (string, error) {
	claim := UserClaimsRefresh{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			Issuer:    "echo-authenticator",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jp.cfg.TtlRefresh())),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claim)
	tokenSigned, err := token.SignedString(jp.cfg.PrivateKeyRefresh())
	if err != nil {
		return "", fmt.Errorf("JWTProvider[GenerateRefreshToken]: %w", err)
	}
	return tokenSigned, nil
}

func (jp *JWTProvier) ParseRefreshToken(tokenString string) (UserClaimsRefresh, error) {
	var claim UserClaimsRefresh
	token, err := jwt.ParseWithClaims(tokenString, &claim, func(token *jwt.Token) (any, error) {
		return jp.cfg.PublicKeyRefresh(), nil
	})

	if err != nil {
		return UserClaimsRefresh{}, fmt.Errorf("JWTProvider[GenerateAccessToken]: %w", err)
	}

	if !token.Valid {
		return UserClaimsRefresh{}, errs.NewError(errs.AuthenticationFailure, nil, domain.TokenInvalid)
	}
	return claim, nil
}
