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
	createPostUC usecase.CreatePostUC
	getPostUC    usecase.GetPostUC
}

func NewPostHandler(
	createPostUC usecase.CreatePostUC,
	getPostUC usecase.GetPostUC,
) *PostHandler {
	return &PostHandler{
		createPostUC: createPostUC,
		getPostUC:    getPostUC,
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
