package api

import (
	"github.com/gin-gonic/gin"
)

type CurrentUser struct {
	id       int64
	username string
	email    string
}

func NewCurrentUser(id int64, username string, email string) CurrentUser {
	return CurrentUser{
		id:       id,
		username: username,
		email:    email,
	}
}

func (u *CurrentUser) ID() int64 {
	return u.id
}

func (u *CurrentUser) Username() string {
	return u.username
}

func (u *CurrentUser) Email() string {
	return u.email
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