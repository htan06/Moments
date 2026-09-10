package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/auth/internal/delivery/handler"
	"github.com/htan06/Moments/auth/internal/security"
)

type JWTMiddleware struct {
	jwtProvider *security.JWTProvier
}

func NewJWTMiddleware(jwtProvider *security.JWTProvier) *JWTMiddleware {
	return &JWTMiddleware{
		jwtProvider: jwtProvider,
	}
}

func (jwtm *JWTMiddleware) RequireAccessToken() gin.HandlerFunc {

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.Split(authorization, " ")
		if parts[0] != "Bearer" || len(parts) < 2 {
			return
		}

		accessToken := parts[1]

		claim, err := jwtm.jwtProvider.ParseAccessToken(accessToken)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		currentUser := handler.NewCurrentUser(claim.UserID, claim.Subject, claim.UserStatus, claim.Username)

		c.Set("currentUser", currentUser)
	}
}

func (jwtm *JWTMiddleware) RequireRefreshToken() gin.HandlerFunc {

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.Split(authorization, " ")
		if parts[0] != "Bearer" || len(parts) < 2 {
			return
		}

		refreshToken := parts[1]

		claim, err := jwtm.jwtProvider.ParseRefreshToken(refreshToken)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		currentUser := handler.NewCurrentUser(claim.UserID, claim.Subject, string(claim.UserStatus), "")

		c.Set("currentUser", currentUser)
	}
}
