package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
	"github.com/htan06/echo-messenger-rest-api/internal/module/auth/usecase"
)

type AuthHandler struct {
	registerUsecase          *usecase.RegisterUsecase
	verifyRegisterOTPUsecase *usecase.VerifyRegisterOTPUsecase
	loginPasswordUsecase     *usecase.LoginPasswordUsecase
	changePasswordUsecase    *usecase.ChangePasswordUsecase
	refreshTokenUsecase      *usecase.RefreshTokenUsecase
}

func NewAuthHandler(
	registerUsecase *usecase.RegisterUsecase,
	verifyRegisterOTPUsecase *usecase.VerifyRegisterOTPUsecase,
	loginPasswordUsecase *usecase.LoginPasswordUsecase,
	changePasswordUsecase *usecase.ChangePasswordUsecase,
	refreshTokenUsecase *usecase.RefreshTokenUsecase,
) *AuthHandler {
	return &AuthHandler{
		registerUsecase:          registerUsecase,
		verifyRegisterOTPUsecase: verifyRegisterOTPUsecase,
		loginPasswordUsecase:     loginPasswordUsecase,
		changePasswordUsecase:    changePasswordUsecase,
		refreshTokenUsecase:      refreshTokenUsecase,
	}
}

func (ah *AuthHandler) handleRegisterUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := ah.registerUsecase.Excute(ctx, req.ToRegisterCmd()); err != nil {
		api.HandleError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (ah *AuthHandler) handleVerifyRegisterOTPUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req VerifyRegisterOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	res, err := ah.verifyRegisterOTPUsecase.Excute(ctx, req.ToVerifyRegisterOTPCmd())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
	})
}

func (ah *AuthHandler) handleLoginPasswordUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req LoginPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	res, err := ah.loginPasswordUsecase.Excute(ctx, req.ToLoginPasswordCmd())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
	})
}

func (ah *AuthHandler) handleChangePasswordUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	var req ChangePasswordReq
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

	if err := ah.changePasswordUsecase.Excute(ctx, changePasswordCmd); err != nil {
		api.HandleError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (ah *AuthHandler) handleRefreshTokenUsecase(c *gin.Context) {
	ctx := c.Request.Context()

	currentUser, ok := api.GetCurrentUser(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}

	accessToken, err := ah.refreshTokenUsecase.Excute(ctx, currentUser.ID())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
