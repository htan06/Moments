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
}

func NewUserHandler(
	getProfileUsecase *usecase.GetProfileUsecase,
	updateProfileUsecase *usecase.UpdateProfileUsecase,
) *UserHandler {
	return &UserHandler{
		getProfileUsecase:    getProfileUsecase,
		updateProfileUsecase: updateProfileUsecase,
	}
}

func (uh *UserHandler) HandlerGetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	username := c.Param("username")
	if username == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	profile, err := uh.getProfileUsecase.Excute(ctx, username)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToProfileResponse(&profile))
}

func (uh *UserHandler) HandlerUpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	var req UpdateProfileReq
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
		UserID:    currentuser.ID(),
		Username:  req.Username,
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
		Bio:       req.Bio,
	}

	profile, err := uh.updateProfileUsecase.Excute(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ToProfileResponse(&profile))
}
