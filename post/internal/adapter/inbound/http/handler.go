package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/htan06/Moments/post/internal/application/domain/entity"
	"github.com/htan06/Moments/post/internal/application/domain/vo"
	"github.com/htan06/Moments/post/internal/application/feature"
)

type PostHandler struct {
	createPostUC       *feature.CreatePostUC
	getPostUC          *feature.GetPostUC
	getPostsByCursorUC *feature.GetPostsByCursorUC
	getFollowingFeed   *feature.GetFollowingFeedUC
	deletePostUC       *feature.DeletePostUC

	likePostUC   *feature.LikePostUC
	unlikePostUC *feature.UnlikePostUC

	getUserRepostsUC *feature.GetRepostsUC
	createRepostUC   *feature.CreateRepostUC
	deleteRepostUC   *feature.DeleteRepostUC
}

func NewPostHandler(
	createPostUC *feature.CreatePostUC,
	getPostUC *feature.GetPostUC,
	getPostsByCursorUC *feature.GetPostsByCursorUC,
	getFollowingFeed *feature.GetFollowingFeedUC,
	deletePostUC *feature.DeletePostUC,
	likePostUC *feature.LikePostUC,
	unlikePostUC *feature.UnlikePostUC,
	getRepostsUC *feature.GetRepostsUC,
	createRepostUC *feature.CreateRepostUC,
	deleteRepostUC *feature.DeleteRepostUC,
) *PostHandler {
	return &PostHandler{
		createPostUC:       createPostUC,
		getPostUC:          getPostUC,
		getPostsByCursorUC: getPostsByCursorUC,
		getFollowingFeed:   getFollowingFeed,

		deletePostUC:     deletePostUC,
		likePostUC:       likePostUC,
		unlikePostUC:     unlikePostUC,
		getUserRepostsUC: getRepostsUC,
		createRepostUC:   createRepostUC,
		deleteRepostUC:   deleteRepostUC,
	}
}

func (ph *PostHandler) HandleUploadPost(c *gin.Context) {
	ctx := c.Request.Context()

	currentuser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	var req struct {
		UploadSessionID string                `json:"session_id"`
		Contents        []vo.Content          `json:"contents"`
		Visibility      entity.Visibility     `json:"visibility"`
		AspectRatio     string                `json:"aspect_ratio"`
		MediaUploads    []feature.MediaUpload `json:"media_uploads"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := feature.UploadPostCmd{
		UpLoadSessionID: req.UploadSessionID,
		AuthorID:        currentuser.ID(),
		Contents:        req.Contents,
		Visibility:      req.Visibility,
		AspectRatio:     req.AspectRatio,
		MediaUploads:    req.MediaUploads,
	}

	postID, err := ph.createPostUC.ExecuteUploadPost(ctx, cmd)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"post_id": *postID,
	})
}

func (ph *PostHandler) HandleCreatePostSession(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		MediaCount int `json:"media_count"`
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

	cmd := feature.CreatePostSessionCmd{
		UserID:     currentUser.ID(),
		MediaCount: req.MediaCount,
	}

	res, err := ph.createPostUC.ExecuteCreatePostSession(ctx, cmd)
	if err != nil {
		HandleError(c, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id":    res.SessionID,
		"media_uploads": res.MediaUploads,
	})
}

func (ph *PostHandler) HandleRequestUploadURLs(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		SessionID  uuid.UUID `json:"session_id"`
		MediaCount int       `json:"media_count"`
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

	cmd := feature.RequestUploadURLs{
		SessionID:  req.SessionID,
		AuthorID:   currentUser.ID(),
		MediaCount: req.MediaCount,
	}

	mediaUploads, err := ph.createPostUC.ExecuteRequestUploadURLs(ctx, cmd)
	if err != nil {
		HandleError(c, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"media_uploads": mediaUploads,
	})
}

func (ph *PostHandler) HandleGetPost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	post, err := ph.getPostUC.Execute(ctx, postID)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, post)
}

func (ph *PostHandler) HandleDeletePost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := feature.DeletePostCmd{
		UserID: currentUser.ID(),
		PostID: postID,
	}

	if err := ph.deletePostUC.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}

	c.AbortWithStatus(http.StatusNoContent)
}

func (ph *PostHandler) HandleGetPosts(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var qry feature.GetPostsByCursorQry
	if err := c.ShouldBindQuery(&qry); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	qry.ViewerID = currentUser.ID()

	resp, err := ph.getPostsByCursorUC.Execute(ctx, qry)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (ph *PostHandler) HandleGetFollowingFeed(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var qry feature.GetFollowingFeedQry
	if err := c.ShouldBindQuery(&qry); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	qry.ViewerID = currentUser.ID()

	resp, err := ph.getFollowingFeed.Execute(ctx, qry)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (ph *PostHandler) HandleLikePost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := feature.LikePostCmd{
		UserID: currentUser.ID(),
		PostID: postID,
	}
	if err := ph.likePostUC.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (ph *PostHandler) HandleUnlikePost(c *gin.Context) {
	ctx := c.Request.Context()

	postID, err := strconv.ParseInt(c.Param("postID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := feature.UnlikePostCmd{
		UserID: currentUser.ID(),
		PostID: postID,
	}
	if err := ph.unlikePostUC.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (ph *PostHandler) HandleGetRepostsByUsername(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")
	if username == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	posts, err := ph.getUserRepostsUC.Execute(ctx, username)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, posts)
}

func (ph *PostHandler) HandleCreateRepost(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		PostID int64 `json:"post_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := feature.CreateRepostCmd{
		UserID: currentUser.ID(),
		PostID: req.PostID,
	}
	id, err := ph.createRepostUC.Execute(ctx, cmd)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"repost_id": id})
}

func (ph *PostHandler) HandleDeleteRepost(c *gin.Context) {
	ctx := c.Request.Context()

	repostID, err := strconv.ParseInt(c.Param("repostID"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, exists := GetCurrentUser(c)
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cmd := feature.DeleteRepostCmd{
		UserID:   currentUser.ID(),
		RepostID: repostID,
	}
	if err := ph.deleteRepostUC.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
