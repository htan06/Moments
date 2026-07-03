package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htan06/echo-messenger-rest-api/internal/api"
)

type AuthHandler struct {
	authService *AuthenticationService
}

func NewAuthenticationHandler(authService *AuthenticationService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (ah *AuthHandler) handleRequireOTP(c *gin.Context) {
	ctx := c.Request.Context()

	var req requireOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
	}

	if err := ah.authService.RequireOTP(ctx, req.Email); err != nil {
		api.HandleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (ah *AuthHandler) handleVerifyOTP(c *gin.Context) {
	ctx := c.Request.Context()

	var req VerifyOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	tokenResp, err := ah.authService.VerifyOTP(ctx, req.Email, req.OTP)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

func (ah *AuthHandler) handleRegisterUser(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	tokenResp, err := ah.authService.RegisterUser(ctx, req)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

func (ah *AuthHandler) handleRefreshToken(c *gin.Context) {
	ctx := c.Request.Context()

	cur, exists := api.GetCurrentUser(c)
	if !exists {
		c.Status(http.StatusUnauthorized)
		return
	}

	accessToken, err := ah.authService.RefreshToken(ctx, cur.ID())
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}