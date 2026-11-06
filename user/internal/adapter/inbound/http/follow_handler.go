package http

import (
	"net/http"
	"strconv"
	"github.com/htan06/Moments/user/internal/application/feature"

	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	createFollow   *feature.CreateFollow
	getFollowing   *feature.GetFollowing
	getFollowers   *feature.GetFollowers
	removefollower *feature.Removefollower
	unfollow       *feature.Unfollow
}

func NewFollowHandler(
	createFollow *feature.CreateFollow,
	getFollowing *feature.GetFollowing,
	getFollowers *feature.GetFollowers,
	removefollower *feature.Removefollower,
	unfollow *feature.Unfollow,
) *FollowHandler {
	return &FollowHandler{
		createFollow:   createFollow,
		getFollowing:   getFollowing,
		getFollowers:   getFollowers,
		removefollower: removefollower,
		unfollow:       unfollow,
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

	cmd := feature.CreateFollowCmd{
		FollowerID:  currentUser.ID(),
		FollowingID: req.FollowingID,
	}

	if err := fh.createFollow.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}

	c.Status(http.StatusCreated)
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

	qry := feature.GetFollowingQry{
		Username: username,
		Cursor:   query.Cursor,
		PageSize: query.PageSize,
	}

	res, err := fh.getFollowing.Execute(ctx, qry)
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

	qry := feature.GetFollowersQry{
		Username: username,
		Cursor:   query.Cursor,
		PageSize: query.PageSize,
	}

	res, err := fh.getFollowers.Execute(ctx, qry)
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

	followerID, err := strconv.ParseInt(c.Param("followerID"), 10, 32)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := feature.RemoveFollowCmd{
		FollowerID:  followerID,
		FollowingID: currentUser.ID(),
	}

	if err := fh.removefollower.Execute(ctx, cmd); err != nil {
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

	followingID, err := strconv.ParseInt(c.Param("followingID"), 10, 32)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := feature.UnfollowCmd{
		FollowerID:  currentUser.ID(),
		FollowingID: followingID,
	}

	if err := fh.unfollow.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
