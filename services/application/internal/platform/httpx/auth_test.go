package httpx

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/htan06/Moments/application/internal/platform/token"
)

type fakeVerifier map[string]token.Claims

func (f fakeVerifier) ParseAccess(s string) (token.Claims, error) {
	c, ok := f[s]
	if !ok {
		return token.Claims{}, token.ErrInvalid
	}
	return c, nil
}

func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := NewEngine(slog.New(slog.NewTextHandler(io.Discard, nil)))
	v := fakeVerifier{
		"good":     {UserID: 7, Username: "alice", Status: "ACTIVE"},
		"inactive": {UserID: 8, Username: "bob", Status: "LOCKED"},
	}
	e.GET("/me", Auth(v), func(c *gin.Context) {
		id, ok := UserID(c)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": id})
	})
	return e
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantCode   string
	}{
		{"no header", "", http.StatusUnauthorized, CodeUnauthorized},
		{"wrong scheme", "Basic good", http.StatusUnauthorized, CodeUnauthorized},
		{"empty token", "Bearer ", http.StatusUnauthorized, CodeUnauthorized},
		{"invalid token", "Bearer nope", http.StatusUnauthorized, CodeUnauthorized},
		{"user not active", "Bearer inactive", http.StatusForbidden, CodeForbidden},
		{"valid", "Bearer good", http.StatusOK, ""},
		{"scheme case-insensitive", "bearer good", http.StatusOK, ""},
	}
	e := newTestEngine()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantCode == "" {
				var body struct {
					UserID int64 `json:"user_id"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.UserID != 7 {
					t.Errorf("body = %s", w.Body.String())
				}
				return
			}
			var eb ErrorBody
			if err := json.Unmarshal(w.Body.Bytes(), &eb); err != nil || eb.Code != tt.wantCode {
				t.Errorf("body = %s, want code %s", w.Body.String(), tt.wantCode)
			}
		})
	}
}
