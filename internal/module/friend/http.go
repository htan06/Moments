package friend

import (
	"net/http"
	"strconv"

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

func (fh *FriendHandler) HandlerCreateFriendRequest(c *gin.Context) {
	ctx := c.Request.Context()

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

	fr, err := fh.friendService.CreateFriendRequest(ctx, currUser.ID(), req.ReceiverID)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, fr)
}

func (fh *FriendHandler) HandlerGetSentFriendRequests(c *gin.Context) {
	ctx := c.Request.Context()

	currUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	friendRequests, err := fh.friendService.GetSentFriendRequests(ctx, currUser.ID())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, friendRequests)
}

func (fh *FriendHandler) HandlerGetReceivedFriendRequests(c *gin.Context) {
	ctx := c.Request.Context()

	currUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	friendRequests, err := fh.friendService.GetReceivedFriendRequests(ctx, currUser.ID())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, friendRequests)
}

func (fh *FriendHandler) HandlerGetListFriends(c *gin.Context) {
	ctx := c.Request.Context()

	currUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	friends, err := fh.friendService.GetListFriends(ctx, currUser.ID())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, friends)
}

func (fh *FriendHandler) HandlerAcceptFriendRequest(c *gin.Context) {
	ctx := c.Request.Context()

	currUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	frID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := fh.friendService.AcceptFriendRequest(ctx, currUser.ID(), int64(frID)); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (fh *FriendHandler) HandlerRejectFriendRequest(c *gin.Context) {
	ctx := c.Request.Context()

	currUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	frID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := fh.friendService.RejectFriendRequest(ctx, currUser.ID(), int64(frID)); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (fh *FriendHandler) HandlerCancelFriendRequest(c *gin.Context) {
	ctx := c.Request.Context()

	currUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	frID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := fh.friendService.CancelFriendRequest(ctx, currUser.ID(), int64(frID)); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
