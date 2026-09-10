package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/htan06/Moments/internal/api"
	"github.com/htan06/Moments/post/internal/domain"
	"github.com/htan06/Moments/post/internal/usecase"
)

type PostHandler struct {
	createPostUC   *usecase.CreatePostUC
	getPostUC      *usecase.GetPostUC
	getUserPostsUC *usecase.GetUserPostsUC
	deletePostUC   *usecase.DeletePostUC

	likePostUC   *usecase.LikePostUC
	unlikePostUC *usecase.UnlikePostUC

	getUserRepostsUC *usecase.GetRepostsUC
	createRepostUC   *usecase.CreateRepostUC
	deleteRepostUC   *usecase.DeleteRepostUC
}

func NewPostHandler(
	createPostUC *usecase.CreatePostUC,
	getPostUC *usecase.GetPostUC,
	getUserPostsUC *usecase.GetUserPostsUC,
	deletePostUC *usecase.DeletePostUC,
	likePostUC *usecase.LikePostUC,
	unlikePostUC *usecase.UnlikePostUC,
	getRepostsUC *usecase.GetRepostsUC,
	createRepostUC *usecase.CreateRepostUC,
	deleteRepostUC *usecase.DeleteRepostUC,
) *PostHandler {
	return &PostHandler{
		createPostUC:     createPostUC,
		getPostUC:        getPostUC,
		getUserPostsUC:   getUserPostsUC,
		deletePostUC:     deletePostUC,
		likePostUC:       likePostUC,
		unlikePostUC:     unlikePostUC,
		getUserRepostsUC: getRepostsUC,
		createRepostUC:   createRepostUC,
		deleteRepostUC:   deleteRepostUC,
	}
}

func (ph *PostHandler) HandlerUploadPost(c *gin.Context) {
	ctx := c.Request.Context()

	currentuser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	var req struct {
		UploadSessionID string            `json:"session_id"`
		Contents        []domain.Content  `json:"contents"`
		Visibility      domain.Visibility `json:"visibility"`
		AspectRatio     string            `json:"aspect_ratio"`
		MediaIds        uuid.UUIDs        `json:"media_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.UploadPostCmd{
		UpLoadSessionID: req.UploadSessionID,
		AuthorID:        currentuser.ID(),
		Contents:        req.Contents,
		Visibility:      req.Visibility,
		AspectRatio:     req.AspectRatio,
		MediaIDs:        req.MediaIds,
	}

	postID, err := ph.createPostUC.ExecuteUploadPost(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"post_id": *postID,
	})
}

func (ph *PostHandler) HandlerCreatePostSession(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		MediaCount int `json:"media_count"`
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

	cmd := usecase.CreatePostSessionCmd{
		UserID:     currentUser.ID(),
		MediaCount: req.MediaCount,
	}

	res, err := ph.createPostUC.ExecuteCreatePostSession(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id":    res.SessionID,
		"media_uploads": res.MediaUploads,
	})
}

func (ph *PostHandler) HandlerRequestUploadURLs(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		SessionID  uuid.UUID `json:"session_id"`
		MediaCount int       `json:"media_count"`
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

	cmd := usecase.RequestUploadURLs{
		SessionID:  req.SessionID,
		AuthorID:   currentUser.ID(),
		MediaCount: req.MediaCount,
	}

	mediaUploads, err := ph.createPostUC.ExecuteRequestUploadURLs(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"media_uploads": mediaUploads,
	})
}

func (ph *PostHandler) HandlerGetPost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	post, err := ph.getPostUC.Excute(ctx, postID)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, post)
}

func (ph *PostHandler) HandlerDeletePost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	currentUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := usecase.DeletePostCmd{
		UserID: currentUser.ID(),
		PostID: postID,
	}

	if err := ph.deletePostUC.Execute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

func (ph *PostHandler) HandlerGetPostsByUsername(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")
	if username == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	posts, err := ph.getUserPostsUC.Excute(ctx, username)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (ph *PostHandler) HandlerLikePost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := usecase.LikePostCmd{
		UserID: currentUser.ID(),
		PostID: postID,
	}
	if err := ph.likePostUC.Execute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (ph *PostHandler) HandlerUnlikePost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := usecase.UnlikePostCmd{
		UserID: currentUser.ID(),
		PostID: postID,
	}
	if err := ph.unlikePostUC.Execute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (ph *PostHandler) HandlerGetRepostsByUsername(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")
	if username == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	posts, err := ph.getUserRepostsUC.Execute(ctx, username)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (ph *PostHandler) HandlerCreateRepost(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		PostID int64 `json:"post_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := usecase.CreateRepostCmd{
		UserID: currentUser.ID(),
		PostID: req.PostID,
	}
	id, err := ph.createRepostUC.Execute(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"repost_id": id})
}

func (ph *PostHandler) HandlerDeleteRepost(c *gin.Context) {
	ctx := c.Request.Context()

	repostID, err := strconv.ParseInt(c.Param("repostID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := api.GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := usecase.DeleteRepostCmd{
		UserID:   currentUser.ID(),
		RepostID: repostID,
	}
	if err := ph.deleteRepostUC.Execute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
