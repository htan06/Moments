package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/htan06/Moments/application/internal/platform/token"
	"github.com/htan06/Moments/application/internal/user/command"
)

// Use-case interfaces, satisfied by the command handlers.
type (
	startRegister interface {
		Handle(ctx context.Context, cmd command.StartRegister) (string, error)
	}
	resendRegisterOTP interface {
		Handle(ctx context.Context, cmd command.ResendRegisterOTP) error
	}
	verifyRegisterOTP interface {
		Handle(ctx context.Context, cmd command.VerifyRegisterOTP) error
	}
	completeRegister interface {
		Handle(ctx context.Context, cmd command.CompleteRegister) (token.Pair, error)
	}
	login interface {
		Handle(ctx context.Context, cmd command.Login) (token.Pair, error)
	}
	refreshToken interface {
		Handle(ctx context.Context, cmd command.RefreshToken) (token.Pair, error)
	}
)

// Auth serves registration, login and token refresh. None of these routes
// require a token.
type Auth struct {
	StartRegister     startRegister
	ResendRegisterOTP resendRegisterOTP
	VerifyRegisterOTP verifyRegisterOTP
	CompleteRegister  completeRegister
	Login             login
	RefreshToken      refreshToken
	Log               *slog.Logger
}

func (a *Auth) Register(r gin.IRouter) {
	g := r.Group("/auth")
	g.POST("/register", a.startRegister)
	g.POST("/register/resend-otp", a.resendRegisterOTP)
	g.POST("/register/verify-otp", a.verifyRegisterOTP)
	g.POST("/register/complete", a.completeRegister)
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

func (a *Auth) startRegister(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c)
		return
	}
	id, err := a.StartRegister.Handle(c.Request.Context(), command.StartRegister{Email: req.Email})
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

func (a *Auth) verifyRegisterOTP(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId"`
		OTP       string `json:"otp"`
	}
	if c.ShouldBindJSON(&req) != nil || req.SessionID == "" {
		badRequest(c)
		return
	}
	err := a.VerifyRegisterOTP.Handle(c.Request.Context(), command.VerifyRegisterOTP{SessionID: req.SessionID, OTP: req.OTP})
	if err != nil {
		writeError(c, a.Log, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *Auth) completeRegister(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}
	if c.ShouldBindJSON(&req) != nil || req.SessionID == "" {
		badRequest(c)
		return
	}
	pair, err := a.CompleteRegister.Handle(c.Request.Context(), command.CompleteRegister{
		SessionID: req.SessionID,
		Username:  req.Username,
		Password:  req.Password,
	})
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
