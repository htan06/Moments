package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/htan06/Moments/application/internal/account/command"
	"github.com/htan06/Moments/application/internal/account/query"
	"github.com/htan06/Moments/application/internal/platform/token"
)

// Auth serves registration, login and token refresh. None of these routes
// require a token. Fields are set by the module wiring.
type Auth struct {
	UsernameAvailability *query.UsernameAvailabilityHandler
	StartRegister        *command.StartRegisterHandler
	ResendRegisterOTP    *command.ResendRegisterOTPHandler
	VerifyRegister       *command.VerifyRegisterHandler
	Login                *command.LoginHandler
	RefreshToken         *command.RefreshTokenHandler
	Log                  *slog.Logger
}

func (a *Auth) Register(r gin.IRouter) {
	g := r.Group("/auth")
	g.GET("/username-availability", a.usernameAvailability)
	g.POST("/register", a.startRegister)
	g.POST("/register/resend-otp", a.resendRegisterOTP)
	g.POST("/register/verify", a.verifyRegister)
	g.POST("/login", a.login)
	g.POST("/refresh", a.refresh)
}

type tokenResponse struct {
	AccessToken      string    `json:"accessToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshToken     string    `json:"refreshToken"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}

func writeTokens(c *gin.Context, status int, p token.Pair) {
	c.JSON(status, tokenResponse{
		AccessToken:      p.AccessToken,
		AccessExpiresAt:  p.AccessExpiresAt.UTC(),
		RefreshToken:     p.RefreshToken,
		RefreshExpiresAt: p.RefreshExpiresAt.UTC(),
	})
}

type sessionResponse struct {
	SessionID string `json:"sessionId"`
}

type availabilityResponse struct {
	Available bool `json:"available"`
}

func (a *Auth) usernameAvailability(c *gin.Context) {
	username, ok := c.GetQuery("username")
	if !ok {
		badRequest(c)
		return
	}
	available, err := a.UsernameAvailability.Handle(c.Request.Context(), query.UsernameAvailability{Username: username})
	if err != nil {
		writeError(c, a.Log, err)
		return
	}
	c.JSON(http.StatusOK, availabilityResponse{Available: available})
}

func (a *Auth) startRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c)
		return
	}
	id, err := a.StartRegister.Handle(c.Request.Context(), command.StartRegister{
		Username: req.Username,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(c, a.Log, err)
		return
	}
	c.JSON(http.StatusOK, sessionResponse{SessionID: id})
}

func (a *Auth) resendRegisterOTP(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId"`
	}
	if c.ShouldBindJSON(&req) != nil || req.SessionID == "" {
		badRequest(c)
		return
	}
	if err := a.ResendRegisterOTP.Handle(c.Request.Context(), command.ResendRegisterOTP{SessionID: req.SessionID}); err != nil {
		writeError(c, a.Log, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *Auth) verifyRegister(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId"`
		OTP       string `json:"otp"`
	}
	if c.ShouldBindJSON(&req) != nil || req.SessionID == "" {
		badRequest(c)
		return
	}
	pair, err := a.VerifyRegister.Handle(c.Request.Context(), command.VerifyRegister{SessionID: req.SessionID, OTP: req.OTP})
	if err != nil {
		writeError(c, a.Log, err)
		return
	}
	writeTokens(c, http.StatusCreated, pair)
}

func (a *Auth) login(c *gin.Context) {
	var req struct {
		Identifier string `json:"identifier"` // email or username
		Password   string `json:"password"`
	}
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c)
		return
	}
	pair, err := a.Login.Handle(c.Request.Context(), command.Login{Identifier: req.Identifier, Password: req.Password})
	if err != nil {
		writeError(c, a.Log, err)
		return
	}
	writeTokens(c, http.StatusOK, pair)
}

func (a *Auth) refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if c.ShouldBindJSON(&req) != nil || req.RefreshToken == "" {
		badRequest(c)
		return
	}
	pair, err := a.RefreshToken.Handle(c.Request.Context(), command.RefreshToken{RefreshToken: req.RefreshToken})
	if err != nil {
		writeError(c, a.Log, err)
		return
	}
	writeTokens(c, http.StatusOK, pair)
}
