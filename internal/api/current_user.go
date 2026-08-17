package api

import (
	"github.com/gin-gonic/gin"
)

type CurrentUser struct {
	id     int64
	email  string
	status string
}

func NewCurrentUser(id int64, email string, status string) CurrentUser {
	return CurrentUser{
		id:     id,
		email:  email,
		status: status,
	}
}

func (u *CurrentUser) ID() int64 {
	return u.id
}

func (u *CurrentUser) Email() string {
	return u.email
}

func (u *CurrentUser) Status() string {
	return u.status
}

func GetCurrentUser(c *gin.Context) (CurrentUser, bool) {
	val, exists := c.Get("currentUser")
	if !exists {
		return CurrentUser{}, false
	}

	currentUser, ok := val.(CurrentUser)
	if !ok {
		return CurrentUser{}, false
	}
	return currentUser, true
}
