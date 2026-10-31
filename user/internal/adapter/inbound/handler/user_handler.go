package handler

import (
	"net/http"
	"user-service/internal/application/feature"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	createProfile *feature.CreateProfile
	getProfile    *feature.GetProfile
	updateProfile *feature.UpdateProfile
	changeAvatar  *feature.ChangeAvatar
	search        *feature.Search
}

func NewUserHandler(
	createProfile *feature.CreateProfile,
	getProfile *feature.GetProfile,
	updateProfile *feature.UpdateProfile,
	changeAvatar *feature.ChangeAvatar,
	search *feature.Search,
) *UserHandler {
	return &UserHandler{
		createProfile: createProfile,
		getProfile:    getProfile,
		updateProfile: updateProfile,
		changeAvatar:  changeAvatar,
		search:        search,
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

	cmd := feature.CreateProfileCmd{
		UserID:   currentuser.ID(),
		Name:     req.Name,
		Username: req.Username,
	}

	userID, err := uh.createProfile.Execute(ctx, cmd)
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

	qry := feature.GetProfileQry{
		CurrentUserID:   currentuser.ID(),
		CurrentUsername: currentuser.Username(),
		TargetUsername:  username,
	}
	profile, err := uh.getProfile.Execute(ctx, qry)
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

	cmd := feature.UpdateProfileCmd{
		UserID:   currentuser.ID(),
		Username: req.Username,
		Name:     req.Name,
		Bio:      req.Bio,
	}

	res, err := uh.updateProfile.Execute(ctx, cmd)
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

	url, err := uh.changeAvatar.ExecuteGetUrlUpload(ctx, currentuser.ID())
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

	avatarURL, err := uh.changeAvatar.ExecuteCompletedUpload(ctx, currentuser.ID())
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

	profiles, err := uh.search.Execute(ctx, username)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profiles)
}
