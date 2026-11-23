// Package httpx holds HTTP plumbing shared by every module: engine setup,
// error response shape, request logging, panic recovery and authentication.
package httpx

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorBody is the JSON shape of every error response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

// Error aborts the request with status and an error body.
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, ErrorBody{Code: code, Message: message})
}

// Common error codes used by platform middleware.
const (
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeInternal     = "INTERNAL_ERROR"
)

// NewEngine returns a gin engine with request logging and panic recovery.
func NewEngine(log *slog.Logger) *gin.Engine {
	e := gin.New()
	e.Use(requestLog(log), recovery(log))
	e.NoRoute(func(c *gin.Context) { Error(c, http.StatusNotFound, "NOT_FOUND", "") })
	return e
}

// requestLog logs method, route template, status and latency. It never logs
// headers, query strings or bodies (they may carry tokens or credentials).
func requestLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		log.LogAttrs(c.Request.Context(), slog.LevelInfo, "http request",
			slog.String("method", c.Request.Method),
			slog.String("route", route),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
		)
	}
}

func recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, rec any) {
		log.ErrorContext(c.Request.Context(), "panic recovered",
			slog.String("route", c.FullPath()),
			slog.Any("panic", rec),
		)
		Error(c, http.StatusInternalServerError, CodeInternal, "")
	})
}
