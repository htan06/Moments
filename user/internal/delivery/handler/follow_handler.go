package handler

import (
	"net/http"
	"strconv"
	"user-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	createFollowUsecase   *usecase.CreateFollowUsecase
	getFollowingUsecase   *usecase.GetFollowingUsecase
	getFollowersUsecase   *usecase.GetFollowersUsecase
	removefollowerUsecase *usecase.RemovefollowerUsecase
	unfollowUsecase       *usecase.UnfollowUsecase
}

func NewFollowHandler(
	createFollowUsecase *usecase.CreateFollowUsecase,
	getFollowingUsecase *usecase.GetFollowingUsecase,
	getFollowersUsecase *usecase.GetFollowersUsecase,
	removefollowerUsecase *usecase.RemovefollowerUsecase,
	unfollowUsecase *usecase.UnfollowUsecase,
) *FollowHandler {
	return &FollowHandler{
		createFollowUsecase:   createFollowUsecase,
		getFollowingUsecase:   getFollowingUsecase,
		getFollowersUsecase:   getFollowersUsecase,
		removefollowerUsecase: removefollowerUsecase,
		unfollowUsecase:       unfollowUsecase,
	}
}

func (fh *FollowHandler) HandlerCreateFollow(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		FollowingID int64 `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	cmd := usecase.CreateFollowCmd{
		FollowerID:  currentUser.ID(),
		FollowingID: req.FollowingID,
	}

	id, err := fh.createFollowUsecase.Execute(ctx, cmd)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"follow_id": *id,
	})
}

func (fh *FollowHandler) HandlerGetFollowing(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")

	type Query struct {
		PageSize int32 `form:"page_size"`
		Cursor   int64 `form:"cursor"`
	}

	var query Query
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	qry := usecase.GetFollowingQry{
		Username: username,
		Cursor:   query.Cursor,
		PageSize: query.PageSize,
	}

	res, err := fh.getFollowingUsecase.Execute(ctx, qry)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"following":      res.Following,
		"current_cursor": res.CurrentCursor,
	})
}

func (fh *FollowHandler) HandlerGetFollowers(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")

	type Query struct {
		PageSize int32 `form:"page_size"`
		Cursor   int64 `form:"cursor"`
	}

	var query Query
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	qry := usecase.GetFollowersQry{
		Username: username,
		Cursor:   query.Cursor,
		PageSize: query.PageSize,
	}

	res, err := fh.getFollowersUsecase.Execute(ctx, qry)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"followers":      res.Followers,
		"current_cursor": res.CurrentCursor,
	})
}

func (fh *FollowHandler) HandlerRemoveFollower(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	followID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.RemoveFollowCmd{
		UserID:   currentUser.ID(),
		FollowID: followID,
	}

	if err := fh.removefollowerUsecase.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
func (fh *FollowHandler) HandlerUnfollow(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	followID, err := strconv.ParseInt(c.Param("id"), 10, 32)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.UnfollowCmd{
		UserID:   currentUser.ID(),
		FollowID: followID,
	}

	if err := fh.unfollowUsecase.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
