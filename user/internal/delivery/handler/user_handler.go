package handler

import (
	"net/http"
	"user-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	createProfileUC      *usecase.CreateProfileUC
	getProfileUsecase    *usecase.GetProfileUsecase
	updateProfileUsecase *usecase.UpdateProfileUsecase
	changeAvatarUsecase  *usecase.ChangeAvatarUsecase
	searchUC             *usecase.SearchUC
}

func NewUserHandler(
	createProfileUC *usecase.CreateProfileUC,
	getProfileUsecase *usecase.GetProfileUsecase,
	updateProfileUsecase *usecase.UpdateProfileUsecase,
	changeAvatarUsecase *usecase.ChangeAvatarUsecase,
	searchUC *usecase.SearchUC,
) *UserHandler {
	return &UserHandler{
		createProfileUC:      createProfileUC,
		getProfileUsecase:    getProfileUsecase,
		updateProfileUsecase: updateProfileUsecase,
		changeAvatarUsecase:  changeAvatarUsecase,
		searchUC:             searchUC,
	}
}

func (uh *UserHandler) HandlerCreateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentuser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	cmd := usecase.CreateProfileCmd{
		UserID:   currentuser.ID(),
		Name:     req.Name,
		Username: req.Username,
	}

	userID, err := uh.createProfileUC.Execute(ctx, cmd)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

func (uh *UserHandler) HandlerGetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")
	if username == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	currentuser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	qry := usecase.GetProfileQry{
		CurrentUserID:   currentuser.ID(),
		CurrentUsername: currentuser.Username(),
		TargetUsername:  username,
	}
	profile, err := uh.getProfileUsecase.Execute(ctx, qry)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (uh *UserHandler) HandlerUpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Username *string `json:"username"`
		Name     *string `json:"name"`
		Bio      *string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentuser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	cmd := usecase.UpdateProfileCmd{
		UserID:   currentuser.ID(),
		Username: req.Username,
		Name:     req.Name,
		Bio:      req.Bio,
	}

	res, err := uh.updateProfileUsecase.Execute(ctx, cmd)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (uh *UserHandler) HandlerGetUrlUploadAvatar(c *gin.Context) {
	ctx := c.Request.Context()

	currentuser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	url, err := uh.changeAvatarUsecase.ExecuteGetUrlUpload(ctx, currentuser.ID())
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"upload_avatar_url": url})
}

func (uh *UserHandler) HandlerCompletedUpload(c *gin.Context) {
	ctx := c.Request.Context()

	currentuser, exists := GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	avatarURL, err := uh.changeAvatarUsecase.ExecuteCompletedUpload(ctx, currentuser.ID())
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": avatarURL})
}

func (uh *UserHandler) HandleFindProfiles(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusOK, nil)
		return
	}

	profiles, err := uh.searchUC.Execute(ctx, username)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profiles)
}
