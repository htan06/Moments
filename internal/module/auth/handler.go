package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/internal/api"
	"github.com/htan06/Moments/internal/config"
	"github.com/htan06/Moments/internal/module/auth/usecase"
)

type AuthHandler struct {
	registerUsecase          *usecase.RegisterUsecase
	verifyRegisterOTPUsecase *usecase.VerifyRegisterOTPUsecase
	loginPasswordUsecase     *usecase.LoginPasswordUsecase
	changePasswordUsecase    *usecase.ChangePasswordUsecase
	refreshTokenUsecase      *usecase.RefreshTokenUsecase
	jwtConfig                *config.JWTConfig
}

func NewAuthHandler(
	registerUsecase *usecase.RegisterUsecase,
	verifyRegisterOTPUsecase *usecase.VerifyRegisterOTPUsecase,
	loginPasswordUsecase *usecase.LoginPasswordUsecase,
	changePasswordUsecase *usecase.ChangePasswordUsecase,
	refreshTokenUsecase *usecase.RefreshTokenUsecase,
	jwtConfig *config.JWTConfig,
) *AuthHandler {
	return &AuthHandler{
		registerUsecase:          registerUsecase,
		verifyRegisterOTPUsecase: verifyRegisterOTPUsecase,
		loginPasswordUsecase:     loginPasswordUsecase,
		changePasswordUsecase:    changePasswordUsecase,
		refreshTokenUsecase:      refreshTokenUsecase,
		jwtConfig:                jwtConfig,
	}
}

func (ah *AuthHandler) handleRegisterUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.RegisterCmd{
		Email:    req.Email,
		Name:     req.Name,
		Username: req.Username,
		Password: req.Password,
	}

	if err := ah.registerUsecase.Execute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (ah *AuthHandler) handleVerifyRegisterOTPUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.VerifyRegisterOTPCmd{
		Email: req.Email,
		OTP:   req.OTP,
	}
	if err := ah.verifyRegisterOTPUsecase.Execute(ctx, cmd); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (ah *AuthHandler) handleLoginPasswordUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.LoginPasswordCmd{
		Email:    req.Email,
		Password: req.Password,
	}

	res, err := ah.loginPasswordUsecase.Execute(ctx, cmd)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       res.UserID,
		"username":      res.Username,
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
	})
}

func (ah *AuthHandler) handleChangePasswordUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, ok := api.GetCurrentUser(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}

	changePasswordCmd := usecase.ChangePasswordCmd{
		Email:           currentUser.Email(),
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	if err := ah.changePasswordUsecase.Execute(ctx, changePasswordCmd); err != nil {
		api.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (ah *AuthHandler) handleRefreshTokenUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, ok := api.GetCurrentUser(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}

	accessToken, err := ah.refreshTokenUsecase.Execute(ctx, currentUser.ID())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
