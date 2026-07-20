package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
	"github.com/htan06/echo-messenger-rest-api/internal/module/user/usecase"
)

type UserHandler struct {
	getProfileUsecase    *usecase.GetProfileUsecase
	updateProfileUsecase *usecase.UpdateProfileUsecase
	changeAvatarUsecase  *usecase.ChangeAvatarUsecase
}

func NewUserHandler(
	getProfileUsecase *usecase.GetProfileUsecase,
	updateProfileUsecase *usecase.UpdateProfileUsecase,
	changeAvatarUsecase *usecase.ChangeAvatarUsecase,
) *UserHandler {
	return &UserHandler{
		getProfileUsecase:    getProfileUsecase,
		updateProfileUsecase: updateProfileUsecase,
		changeAvatarUsecase:  changeAvatarUsecase,
	}
}

func (uh *UserHandler) HandlerGetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")
	if username == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	profile, err := uh.getProfileUsecase.Execute(ctx, username)
	if err != nil {
		api.HandleError(c, err)
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

	currentuser, exists := api.GetCurrentUser(c)
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

	if err := uh.updateProfileUsecase.Execute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (uh *UserHandler) HandlerGetUrlUploadAvatar(c *gin.Context) {
	ctx := c.Request.Context()

	currentuser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	url, err := uh.changeAvatarUsecase.ExecuteGetUrlUpload(ctx, currentuser.ID())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"upload_avatar_url": url})
}

func (uh *UserHandler) HandlerCompletedUpload(c *gin.Context) {
	ctx := c.Request.Context()

	currentuser, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	if err := uh.changeAvatarUsecase.ExecuteCompletedUpload(ctx, currentuser.ID()); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}
