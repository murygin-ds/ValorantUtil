package riot_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ValorantAPI/internal/api/v1/riot"
	"ValorantAPI/internal/config"
	"ValorantAPI/internal/deps"
	"ValorantAPI/internal/logger"
	redisstorage "ValorantAPI/internal/storage/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testLogger() *logger.Logger {
	return &logger.Logger{SugaredLogger: zap.NewNop().Sugar()}
}

func buildRiotDeps(t *testing.T) (*deps.Deps, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return &deps.Deps{
		Cfg:             &config.Config{Security: config.SecurityConfig{Secret: "test-secret"}},
		Logging:         testLogger(),
		PendingAuthRepo: redisstorage.NewPendingAuthRepo(rdb),
		SessionRepo:     redisstorage.NewSessionRepo(rdb),
		HTTPClient:      &http.Client{},
	}, mr
}

func newRiotRouter(d *deps.Deps) *gin.Engine {
	r := gin.New()
	h := riot.NewHandler(d)

	auth := func(c *gin.Context) { c.Set("user_id", int64(1)); c.Next() }

	r.GET("/auth/url", auth, h.GetAuthURL)
	r.POST("/callback", auth, h.RiotAuthCallback)
	r.POST("/login", auth, h.RiotLogin)
	r.POST("/login/mfa", auth, h.RiotLoginMFA)
	r.POST("/login/captcha", auth, h.RiotLoginCaptcha)
	r.POST("/auth/submit-url", auth, h.SubmitRedirectURL)
	return r
}

func TestGetAuthURL_ReturnsURL(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/auth/url", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	authURL, ok := resp["auth_url"].(string)
	require.True(t, ok)
	assert.Contains(t, authURL, "auth.riotgames.com")
}

func TestRiotAuthCallback_InvalidBody_Returns400(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRiotAuthCallback_MissingRequiredField_Returns400(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	body := jsonBody(t, map[string]string{"id_token": "only_id"})
	req := httptest.NewRequest(http.MethodPost, "/callback", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRiotLogin_InvalidBody_Returns400(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{bad}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRiotLoginMFA_InvalidBody_Returns400(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/login/mfa", bytes.NewBufferString("{bad}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRiotLoginMFA_SessionNotFound_Returns401(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	body := jsonBody(t, map[string]string{"session_id": "no", "code": "123456"})
	req := httptest.NewRequest(http.MethodPost, "/login/mfa", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRiotLoginCaptcha_InvalidBody_Returns400(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/login/captcha", bytes.NewBufferString("{bad}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRiotLoginCaptcha_SessionNotFound_Returns401(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	body := jsonBody(t, map[string]string{
		"session_id":    "no",
		"username":      "user",
		"password":      "pass",
		"captcha_token": "tok",
	})
	req := httptest.NewRequest(http.MethodPost, "/login/captcha", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSubmitRedirectURL_InvalidBody_Returns400(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/auth/submit-url", bytes.NewBufferString("{bad}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitRedirectURL_MissingURL_Returns400(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	body := jsonBody(t, map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/auth/submit-url", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubmitRedirectURL_NoTokenInURL_Returns400(t *testing.T) {
	d, _ := buildRiotDeps(t)
	r := newRiotRouter(d)

	body := jsonBody(t, map[string]string{"redirect_url": "https://playvalorant.com/opt_in#state=test"})
	req := httptest.NewRequest(http.MethodPost, "/auth/submit-url", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}
