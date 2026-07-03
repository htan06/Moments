package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
)

type UserHandler struct {
	userServie *UserService
}

func NewUserHandler(userServie *UserService) *UserHandler {
	return &UserHandler{
		userServie: userServie,
	}
}

func (uh *UserHandler) HandleGetCurrentUserProfile(c *gin.Context) {
	ctx := c.Request.Context()

	cur, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	profile, err := uh.userServie.GetProfileById(ctx, cur.ID())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (uh *UserHandler) HandleUpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	cur, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	var req UpdateInfoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := uh.userServie.UpdateInfo(ctx, cur.ID(), req); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (uh *UserHandler) HandleChangeReadStatus(c *gin.Context) {
	ctx := c.Request.Context()

	cur, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	var req ChangeReadStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := uh.userServie.ChangeReadStatus(ctx, cur.ID(), req); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (uh *UserHandler) HandleUpdateUsername(c *gin.Context) {
	ctx := c.Request.Context()

	cur, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	var req UpdateUsernameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := uh.userServie.UpdateUsername(ctx, cur.ID(), req); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (uh *UserHandler) HandlerFindUserByUsername(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")

	profile, err := uh.userServie.GetProfileByUserName(ctx, username)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}
