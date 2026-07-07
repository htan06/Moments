package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
	"github.com/htan06/echo-messenger-rest-api/internal/security"
)

type JWTMiddleWare struct {
	jwtProvider *security.JWTProvier
}

func NewJWTMiddleware(jwtProvider *security.JWTProvier) *JWTMiddleWare {
	return &JWTMiddleWare{
		jwtProvider: jwtProvider,
	}
}

func (jwtm *JWTMiddleWare) RequireAccessToken() gin.HandlerFunc {

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.Split(authorization, " ")
		if parts[0] != "Bearer" {
			return
		}

		accessToken := parts[1]

		claim, err := jwtm.jwtProvider.ParseAccessToken(accessToken)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		currentUser := api.NewCurrentUser(claim.UserID, claim.Username, claim.Subject)

		c.Set("currentUser", currentUser)
	}
}

func (jwtm *JWTMiddleWare) RequireRefreshToken() gin.HandlerFunc {

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.Split(authorization, " ")
		if parts[0] != "Bearer" {
			return
		}

		refreshToken := parts[1]

		claim, err := jwtm.jwtProvider.ParseRefreshToken(refreshToken)
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		currentUser := api.NewCurrentUser(claim.UserID, claim.Username, claim.Subject)

		c.Set("currentUser", currentUser)
	}
}