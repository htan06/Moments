// Package httpapi exposes the user module over HTTP. It is the only place that
// knows error codes and HTTP status codes.
package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/htan06/Moments/application/internal/platform/httpx"
	"github.com/htan06/Moments/application/internal/user/domain"
)

// Error codes of the user module.
const (
	codeInvalidRequest      = "INVALID_REQUEST"
	codeInvalidEmail        = "INVALID_EMAIL"
	codeInvalidUsername     = "INVALID_USERNAME"
	codeInvalidPassword     = "INVALID_PASSWORD"
	codeUsernameTaken       = "USERNAME_TAKEN"
	codeRegistrationFailed  = "REGISTRATION_FAILED"
	codeInvalidCredentials  = "INVALID_CREDENTIALS"
	codeSessionNotFound     = "SESSION_NOT_FOUND"
	codeSessionNotVerified  = "SESSION_NOT_VERIFIED"
	codeSessionVerified     = "SESSION_ALREADY_VERIFIED"
	codeWrongOTP            = "WRONG_OTP"
	codeTooManyOTPAttempts  = "TOO_MANY_OTP_ATTEMPTS"
	codeOTPResendTooSoon    = "OTP_RESEND_TOO_SOON"
	codeTooManyAttempts     = "TOO_MANY_ATTEMPTS"
	codeInvalidRefreshToken = "INVALID_REFRESH_TOKEN"
)

type errorMapping struct {
	err    error
	status int
	code   string
}

var mappings = []errorMapping{
	{domain.ErrInvalidEmail, http.StatusBadRequest, codeInvalidEmail},
	{domain.ErrInvalidUsername, http.StatusBadRequest, codeInvalidUsername},
	{domain.ErrInvalidPassword, http.StatusBadRequest, codeInvalidPassword},
	{domain.ErrUsernameTaken, http.StatusConflict, codeUsernameTaken},
	{domain.ErrRegistrationFailed, http.StatusConflict, codeRegistrationFailed},
	{domain.ErrInvalidCredentials, http.StatusUnauthorized, codeInvalidCredentials},
	{domain.ErrSessionNotFound, http.StatusNotFound, codeSessionNotFound},
	{domain.ErrSessionNotVerified, http.StatusConflict, codeSessionNotVerified},
	{domain.ErrSessionVerified, http.StatusConflict, codeSessionVerified},
	{domain.ErrWrongOTP, http.StatusBadRequest, codeWrongOTP},
	{domain.ErrTooManyOTPAttempts, http.StatusBadRequest, codeTooManyOTPAttempts},
	{domain.ErrInvalidToken, http.StatusUnauthorized, codeInvalidRefreshToken},
	{domain.ErrAccountInactive, http.StatusForbidden, httpx.CodeForbidden},
}

// writeError maps a use-case error to a response. Unknown errors are logged
// and returned as 500 without details.
func writeError(c *gin.Context, log *slog.Logger, err error) {
	var resend *domain.ResendTooSoonError
	if errors.As(err, &resend) {
		tooMany(c, codeOTPResendTooSoon, resend.RetryAfterSeconds)
		return
	}
	var blocked *domain.TooManyAttemptsError
	if errors.As(err, &blocked) {
		tooMany(c, codeTooManyAttempts, blocked.RetryAfterSeconds)
		return
	}
	for _, m := range mappings {
		if errors.Is(err, m.err) {
			httpx.Error(c, m.status, m.code, "")
			return
		}
	}
	log.ErrorContext(c.Request.Context(), "user request failed",
		slog.String("route", c.FullPath()), slog.Any("err", err))
	httpx.Error(c, http.StatusInternalServerError, httpx.CodeInternal, "")
}

// tooMany answers 429 with the remaining wait in Retry-After and in the body
// (USER-R3: the response tells how long is left).
func tooMany(c *gin.Context, code string, seconds int) {
	c.Header("Retry-After", strconv.Itoa(seconds))
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"code":              code,
		"retryAfterSeconds": seconds,
	})
}

func badRequest(c *gin.Context) {
	httpx.Error(c, http.StatusBadRequest, codeInvalidRequest, "")
}
