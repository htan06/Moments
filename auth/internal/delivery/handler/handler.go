package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/Moments/auth/config"
	"github.com/htan06/Moments/auth/internal/usecase"
)

type AuthHandler struct {
	activeUserUC             *usecase.ActiveUserUC
	registerUsecase          *usecase.RegisterUsecase
	verifyRegisterOTPUsecase *usecase.VerifyRegisterOTPUsecase
	loginPasswordUsecase     *usecase.LoginPasswordUsecase
	changePasswordUsecase    *usecase.ChangePasswordUsecase
	refreshTokenUsecase      *usecase.RefreshTokenUsecase
	jwtConfig                *config.JWTConfig
}

func NewAuthHandler(
	activeUserUC *usecase.ActiveUserUC,
	registerUsecase *usecase.RegisterUsecase,
	verifyRegisterOTPUsecase *usecase.VerifyRegisterOTPUsecase,
	loginPasswordUsecase *usecase.LoginPasswordUsecase,
	changePasswordUsecase *usecase.ChangePasswordUsecase,
	refreshTokenUsecase *usecase.RefreshTokenUsecase,
	jwtConfig *config.JWTConfig,
) *AuthHandler {
	return &AuthHandler{
		activeUserUC:             activeUserUC,
		registerUsecase:          registerUsecase,
		verifyRegisterOTPUsecase: verifyRegisterOTPUsecase,
		loginPasswordUsecase:     loginPasswordUsecase,
		changePasswordUsecase:    changePasswordUsecase,
		refreshTokenUsecase:      refreshTokenUsecase,
		jwtConfig:                jwtConfig,
	}
}

func (ah *AuthHandler) HandleRegisterUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cmd := usecase.RegisterCmd{
		Email:    req.Email,
		Password: req.Password,
	}

	if err := ah.registerUsecase.Execute(ctx, cmd); err != nil {
		HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (ah *AuthHandler) HandleActiveUserUC(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, ok := GetCurrentUser(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}

	res, err := ah.activeUserUC.Execute(ctx, currentUser.ID())
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     res.UserID,
		"user_status": res.Status,
		"username":    res.Username,
		"tokens": gin.H{
			"access_token":  res.AccessToken,
			"refresh_token": res.RefreshToken,
		},
	})
}

func (ah *AuthHandler) HandleVerifyRegisterOTPUsecase(c *gin.Context) {
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
		HandleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (ah *AuthHandler) HandleLoginPasswordUsecase(c *gin.Context) {
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
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     res.UserID,
		"user_status": res.Status,
		"username":    res.Username,
		"tokens": gin.H{
			"access_token":  res.AccessToken,
			"refresh_token": res.RefreshToken,
		},
	})
}

func (ah *AuthHandler) HandleChangePasswordUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	currentUser, ok := GetCurrentUser(c)
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
		HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (ah *AuthHandler) HandleRefreshTokenUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, ok := GetCurrentUser(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}

	accessToken, err := ah.refreshTokenUsecase.Execute(ctx, currentUser.ID())
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
