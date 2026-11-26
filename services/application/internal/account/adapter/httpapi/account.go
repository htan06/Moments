package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/htan06/Moments/application/internal/account/command"
	"github.com/htan06/Moments/application/internal/platform/httpx"
)

// Account serves the signed-in user's own account: password and username.
// Every route requires an access token; the user id is taken from it, never
// from the request. Fields are set by the module wiring.
type Account struct {
	ChangePassword *command.ChangePasswordHandler
	ChangeUsername *command.ChangeUsernameHandler
	// RequireAuth is httpx.Auth with the token verifier.
	RequireAuth gin.HandlerFunc
	Log         *slog.Logger
}

func (a *Account) Register(r gin.IRouter) {
	g := r.Group("/account", a.RequireAuth)
	g.PUT("/password", a.changePassword)
	g.PUT("/username", a.changeUsername)
}

func (a *Account) changePassword(c *gin.Context) {
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "")
		return
	}
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c)
		return
	}
	err := a.ChangePassword.Handle(c.Request.Context(), command.ChangePassword{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		writeError(c, a.Log, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *Account) changeUsername(c *gin.Context) {
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, httpx.CodeUnauthorized, "")
		return
	}
	var req struct {
		Username string `json:"username"`
	}
	if c.ShouldBindJSON(&req) != nil {
		badRequest(c)
		return
	}
	pair, err := a.ChangeUsername.Handle(c.Request.Context(), command.ChangeUsername{UserID: userID, Username: req.Username})
	if err != nil {
		writeError(c, a.Log, err)
		return
	}
	writeTokens(c, http.StatusOK, pair)
}
