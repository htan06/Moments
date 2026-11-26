// Package httpapi exposes the account module over HTTP. It is the only place that
// knows error codes and HTTP status codes.
package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/htan06/Moments/application/internal/account/domain"
	"github.com/htan06/Moments/application/internal/platform/httpx"
)

// Error codes of the account module.
const (
	codeInvalidRequest      = "INVALID_REQUEST"
	codeInvalidEmail        = "INVALID_EMAIL"
	codeInvalidUsername     = "INVALID_USERNAME"
	codeInvalidPassword     = "INVALID_PASSWORD"
	codeInvalidName         = "INVALID_NAME"
	codeUsernameTaken       = "USERNAME_TAKEN"
	codeRegistrationFailed  = "REGISTRATION_FAILED"
	codeInvalidCredentials  = "INVALID_CREDENTIALS"
	codeSessionNotFound     = "SESSION_NOT_FOUND"
	codeWrongOTP            = "WRONG_OTP"
	codeTooManyOTPAttempts  = "TOO_MANY_OTP_ATTEMPTS"
	codeOTPResendTooSoon    = "OTP_RESEND_TOO_SOON"
	codeTooManyAttempts     = "TOO_MANY_ATTEMPTS"
	codeInvalidRefreshToken = "INVALID_REFRESH_TOKEN"
	codeWrongPassword       = "WRONG_PASSWORD"
	codeSamePassword        = "SAME_PASSWORD"
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
	{domain.ErrInvalidName, http.StatusBadRequest, codeInvalidName},
	{domain.ErrUsernameTaken, http.StatusConflict, codeUsernameTaken},
	{domain.ErrRegistrationFailed, http.StatusConflict, codeRegistrationFailed},
	{domain.ErrInvalidCredentials, http.StatusUnauthorized, codeInvalidCredentials},
	{domain.ErrSessionNotFound, http.StatusNotFound, codeSessionNotFound},
	{domain.ErrWrongOTP, http.StatusBadRequest, codeWrongOTP},
	{domain.ErrTooManyOTPAttempts, http.StatusBadRequest, codeTooManyOTPAttempts},
	{domain.ErrInvalidToken, http.StatusUnauthorized, codeInvalidRefreshToken},
	{domain.ErrAccountInactive, http.StatusForbidden, httpx.CodeForbidden},
	{domain.ErrWrongPassword, http.StatusBadRequest, codeWrongPassword},
	{domain.ErrSamePassword, http.StatusBadRequest, codeSamePassword},
	// Only reached by signed-in routes: the token outlived its account.
	{domain.ErrAccountNotFound, http.StatusUnauthorized, httpx.CodeUnauthorized},
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
// (ACCOUNT-R3: the response tells how long is left).
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
