package friend

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
)

type FriendHandler struct {
	friendService *FriendService
}

func NewFriendHandler(friendService *FriendService) *FriendHandler {
	return &FriendHandler{
		friendService: friendService,
	}
}

func (fh *FriendHandler) HandlerFindUserByUsername(c *gin.Context) {
	ctx := c.Request.Context()
	
	username := c.Param("username")

	up, err := fh.friendService.FindUserByUserName(ctx, username)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, up)
}

func (fh *FriendHandler) HandlerFriendRequest(c *gin.Context) {
	ctx :=  c.Request.Context()

	var req FriendRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	if err := fh.friendService.CreateFriendRequest(ctx, currUser.ID(), req.ReceiverID); err != nil {
		api.HandleError(c, err)
		return
	}
	c.Status(http.StatusOK)
}