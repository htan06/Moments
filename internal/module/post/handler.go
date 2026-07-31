package post

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/usecase"
)

type PostHandler struct {
	createPostUC   usecase.CreatePostUC
	getPostUC      usecase.GetPostUC
	getUserPostsUC usecase.GetUserPostsUC
	deletePostUC   usecase.DeletePostUC
}

func NewPostHandler(
	createPostUC usecase.CreatePostUC,
	getPostUC usecase.GetPostUC,
	getUserPostsUC usecase.GetUserPostsUC,
	deletePostUC usecase.DeletePostUC,
) *PostHandler {
	return &PostHandler{
		createPostUC:   createPostUC,
		getPostUC:      getPostUC,
		getUserPostsUC: getUserPostsUC,
		deletePostUC:   deletePostUC,
	}
}

func (ph *PostHandler) handlerCreatePost(c *gin.Context) {
	ctx := c.Request.Context()

	currentuser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	var req struct {
		UploadSessionID string             `json:"upload_session_id"`
		Contents        []usecase.Content  `json:"contents"`
		Visibility      domain.Visibility  `json:"visibility"`
		AspectRatio     domain.AspectRatio `json:"aspect_ratio"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.CreatePostCmd{
		UpLoadSessionID: req.UploadSessionID,
		AuthorID:        currentuser.ID(),
		Contents:        req.Contents,
		Visibility:      req.Visibility,
		AspectRatio:     req.AspectRatio,
	}

	postID, err := ph.createPostUC.ExecuteCreatePost(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"post_id": *postID,
	})
}

func (ph *PostHandler) handlerPrepareUploadPost(c *gin.Context) {
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

	cmd := usecase.PrepareUploadPostCmd{
		UserID:     currentUser.ID(),
		MediaCount: req.MediaCount,
	}

	res, err := ph.createPostUC.ExecutePrepareUploadPost(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"upload_session_id": res.SessionID,
		"presigned_urls":    res.PresignedURLs,
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
