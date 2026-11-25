package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/auth/internal/delivery/handler"
)

func RequireUserStatus(status string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, ok := handler.GetCurrentUser(c)
		if !ok || currentUser.Status() != status {
			c.Status(http.StatusUnauthorized)
			c.Abort()
		}
	}
}
