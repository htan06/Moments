package post

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/htan06/Moments/internal/api"
	"github.com/htan06/Moments/internal/module/post/domain"
	"github.com/htan06/Moments/internal/module/post/usecase"
)

type PostHandler struct {
	createPostUC   usecase.CreatePostUC
	getPostUC      usecase.GetPostUC
	getUserPostsUC usecase.GetUserPostsUC
	deletePostUC   usecase.DeletePostUC

	likePostUC   usecase.LikePostUC
	unlikePostUC usecase.UnlikePostUC
}

func NewPostHandler(
	createPostUC usecase.CreatePostUC,
	getPostUC usecase.GetPostUC,
	getUserPostsUC usecase.GetUserPostsUC,
	deletePostUC usecase.DeletePostUC,
	likePostUC usecase.LikePostUC,
	unlikePostUC usecase.UnlikePostUC,
) *PostHandler {
	return &PostHandler{
		createPostUC:   createPostUC,
		getPostUC:      getPostUC,
		getUserPostsUC: getUserPostsUC,
		deletePostUC:   deletePostUC,
		likePostUC:     likePostUC,
		unlikePostUC:   unlikePostUC,
	}
}

func (ph *PostHandler) handlerUploadPost(c *gin.Context) {
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

func (ph *PostHandler) handlerCreatePostSession(c *gin.Context) {
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

func (ph *PostHandler) handlerRequestUploadURLs(c *gin.Context) {
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

func (ph *PostHandler) handlerGetPost(c *gin.Context) {
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

func (ph *PostHandler) handlerDeletePost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
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

func (ph *PostHandler) handlerGetPostsByUsername(c *gin.Context) {
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

func (ph *PostHandler) handlerLikePost(c *gin.Context) {
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

func (ph *PostHandler) handlerUnlikePost(c *gin.Context) {
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

	c.Status(http.StatusOK)
}
