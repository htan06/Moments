package httpx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/htan06/Moments/application/internal/platform/token"
)

// AccessVerifier verifies an access token.
type AccessVerifier interface {
	ParseAccess(s string) (token.Claims, error)
}

// statusActive is the only user_status allowed to call authenticated APIs.
const statusActive = "ACTIVE"

const claimsKey = "httpx.claims"

// Auth requires a valid `Authorization: Bearer <access token>` header from an
// ACTIVE user. The user identity is then available via UserID / CurrentClaims;
// handlers must never take it from the body or query.
func Auth(v AccessVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := bearer(c.GetHeader("Authorization"))
		if !ok {
			Error(c, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}
		claims, err := v.ParseAccess(raw)
		if err != nil {
			Error(c, http.StatusUnauthorized, CodeUnauthorized, "")
			return
		}
		if claims.Status != statusActive {
			Error(c, http.StatusForbidden, CodeForbidden, "")
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// UserID returns the authenticated user's id. ok is false outside Auth.
func UserID(c *gin.Context) (id int64, ok bool) {
	cl, ok := CurrentClaims(c)
	return cl.UserID, ok
}

// CurrentClaims returns the authenticated user's claims. ok is false outside Auth.
func CurrentClaims(c *gin.Context) (token.Claims, bool) {
	v, ok := c.Get(claimsKey)
	if !ok {
		return token.Claims{}, false
	}
	cl, ok := v.(token.Claims)
	return cl, ok
}

func bearer(h string) (string, bool) {
	scheme, tok, ok := strings.Cut(h, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	tok = strings.TrimSpace(tok)
	return tok, tok != ""
}
