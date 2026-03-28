package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ValorantAPI/internal/middleware"
	"ValorantAPI/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const secret = "test-secret"

func init() {
	gin.SetMode(gin.TestMode)
}

func newAuthRouter() (*gin.Engine, *int64) {
	r := gin.New()
	var capturedUserID int64
	r.GET("/protected", middleware.Auth(secret), func(c *gin.Context) {
		capturedUserID = c.GetInt64("user_id")
		c.Status(http.StatusOK)
	})
	return r, &capturedUserID
}

func TestAuth_NoCookie_Returns401(t *testing.T) {
	r, _ := newAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	r, _ := newAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid.jwt.token"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ValidToken_SetsUserID(t *testing.T) {
	token, err := jwt.Generate(99, secret, time.Hour)
	require.NoError(t, err)

	r, capturedID := newAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(99), *capturedID)
}

func TestAuth_ExpiredToken_Returns401(t *testing.T) {
	token, err := jwt.Generate(1, secret, -time.Second)
	require.NoError(t, err)

	r, _ := newAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_WrongSecret_Returns401(t *testing.T) {
	token, err := jwt.Generate(1, "other-secret", time.Hour)
	require.NoError(t, err)

	r, _ := newAuthRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_RequestProceedsAfterMiddleware(t *testing.T) {
	token, err := jwt.Generate(7, secret, time.Hour)
	require.NoError(t, err)

	r := gin.New()
	var reached bool
	r.GET("/protected", middleware.Auth(secret), func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.True(t, reached)
}
