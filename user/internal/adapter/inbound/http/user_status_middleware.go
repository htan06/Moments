package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/user/internal/domain"
)

func requireUserStatus(status string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUser, ok := GetCurrentUser(c)
		if !ok || currentUser.Status() != status {
			c.Status(http.StatusUnauthorized)
			c.Abort()
		}
	}
}

func RequireUserActive() gin.HandlerFunc {
	return requireUserStatus(domain.UserActive)
}

func RequireUserPending() gin.HandlerFunc {
	return requireUserStatus(domain.UserPending)
}