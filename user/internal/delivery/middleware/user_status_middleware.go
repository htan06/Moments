package middleware

import (
	"net/http"
	"user-service/internal/delivery/handler"

	"github.com/gin-gonic/gin"
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
