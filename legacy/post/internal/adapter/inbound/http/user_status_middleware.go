package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	userActive  = "ACTIVE"
	userPending = "PENDING"
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
	return requireUserStatus(userActive)
}

func RequireUserPending() gin.HandlerFunc {
	return requireUserStatus(userPending)
}
