package follow

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
	"github.com/htan06/echo-messenger-rest-api/internal/module/follow/usecase"
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

func (fh *FollowHandler) handlerCreateFollow(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		FollowingID int64 `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	cmd := usecase.CreateFollowCmd{
		FollowerID:  currentUser.ID(),
		FollowingID: req.FollowingID,
	}

	id, err := fh.createFollowUsecase.Excute(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"follow_id": *id,
	})
}

func (fh *FollowHandler) handlerGetFollowing(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	qry := usecase.GetFollowingQry{
		UserID:   userID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	users, err := fh.getFollowingUsecase.Excute(ctx, qry)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (fh *FollowHandler) handlerGetFollowers(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	page, err := strconv.ParseInt(c.Query("page"), 10, 32)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	pageSize, err := strconv.ParseInt(c.Query("page_size"), 10, 32)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	qry := usecase.GetFollowersQry{
		UserID:   userID,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	users, err := fh.getFollowersUsecase.Excute(ctx, qry)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (fh *FollowHandler) handlerRemoveFollower(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, exists := api.GetCurrentUser(c)
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

	if err := fh.removefollowerUsecase.Excute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
func (fh *FollowHandler) handlerUnfollow(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, exists := api.GetCurrentUser(c)
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

	if err := fh.unfollowUsecase.Excute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
