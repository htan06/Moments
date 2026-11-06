package http

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/htan06/Moments/user/config"
	"github.com/htan06/Moments/user/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserClaimsAccess struct {
	UserID     int64  `json:"user_id"`
	UserStatus string `json:"user_status"`
	Username   string `json:"username"`
	jwt.RegisteredClaims
}

type AccessTokenMiddleware struct {
	publicKeyAccess *rsa.PublicKey
}

func NewJWTMiddleware(cfg *config.AuthenticationConfig) *AccessTokenMiddleware {
	return &AccessTokenMiddleware{
		publicKeyAccess: cfg.JWTPublicKey,
	}
}

func (jwtm *AccessTokenMiddleware) Require() gin.HandlerFunc {

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.Split(authorization, " ")
		if parts[0] != "Bearer" || len(parts) < 2 {
			return
		}

		accessToken := parts[1]

		claim, err := jwtm.parseAccessToken(accessToken)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		currentUser := NewCurrentUser(claim.UserID, claim.Subject, string(claim.UserStatus), claim.Username)

		c.Set("currentUser", currentUser)
	}
}

func (jp *AccessTokenMiddleware) parseAccessToken(tokenString string) (UserClaimsAccess, error) {
	var claim UserClaimsAccess
	token, err := jwt.ParseWithClaims(tokenString, &claim, func(token *jwt.Token) (any, error) {
		return jp.publicKeyAccess, nil
	})

	if err != nil {
		return UserClaimsAccess{}, fmt.Errorf("JWTProvider[GenerateAccessToken]: %w", err)
	}

	if !token.Valid {
		return UserClaimsAccess{}, domain.ErrTokenInvalid
	}
	return claim, nil
}
