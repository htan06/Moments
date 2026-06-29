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

func (ah *AuthHandler) handleSendOTP(c *gin.Context) {
	ctx := c.Request.Context()

	var req SendOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
	}

	if err := ah.authService.SendOTP(ctx, req.Email); err != nil {
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
	}

	tokenResp, err := ah.authService.RegisterUser(ctx, req)
	if err != nil {
		api.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}
