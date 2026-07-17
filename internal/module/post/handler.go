package post

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/domain"
	"github.com/htan06/echo-messenger-rest-api/internal/module/post/usecase"
)

type PostHandler struct {
	createPostUC usecase.CreatePostUC
}

func NewPostHandler(createPostUC usecase.CreatePostUC) *PostHandler {
	return &PostHandler{
		createPostUC: createPostUC,
	}
}

func (ph *PostHandler) handlerCreatePost(c *gin.Context) {
	ctx := c.Request.Context()

	currentuser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	type CreatePostRequest struct {
		Contents   []usecase.Content `json:"contents"`
		Visibility domain.Visibility `json:"visibility"`
		MediaCount int               `json:"media_count"`
	}

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.CreatePostCmd{
		AuthorID:    currentuser.ID(),
		Contents:    req.Contents,
		Visibility:  req.Visibility,
		MediaCounts: req.MediaCount,
	}

	res, err := ph.createPostUC.Excute(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"post_session_id": res.PostSessionID,
		"upload_urls":     res.UploadURLs,
	})
}
