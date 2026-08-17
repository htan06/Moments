package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/internal/api"
)

func RequireUserStatus(status string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, ok := api.GetCurrentUser(c)
		if !ok || currentUser.Status() != status {
			c.Status(http.StatusUnauthorized)
			c.Abort()
		}
	}
}
