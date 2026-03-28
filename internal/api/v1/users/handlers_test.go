package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ValorantAPI/internal/api/v1/users"
	"ValorantAPI/internal/config"
	"ValorantAPI/internal/deps"
	"ValorantAPI/internal/domain/user"
	domainvalorant "ValorantAPI/internal/domain/valorant"
	"ValorantAPI/internal/logger"
	"ValorantAPI/internal/pkg/hash"
	"ValorantAPI/internal/storage/postgres"
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

type stubUserRepo struct {
	createErr error
	getErr    error
	stored    user.User
}

func (r *stubUserRepo) CreateUser(_ context.Context, u *user.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	u.ID = r.stored.ID
	return nil
}

func (r *stubUserRepo) GetUserByLogin(_ context.Context, u *user.User) error {
	if r.getErr != nil {
		return r.getErr
	}
	*u = r.stored
	return nil
}

func (r *stubUserRepo) GetUserByID(_ context.Context, u *user.User) error {
	if r.getErr != nil {
		return r.getErr
	}
	*u = r.stored
	return nil
}

type stubValorantRepo struct {
	accounts []domainvalorant.Account
	listErr  error
}

func (r *stubValorantRepo) CreateAccount(_ context.Context, _ *domainvalorant.Account) error {
	return nil
}

func (r *stubValorantRepo) GetAccountsList(_ context.Context, _, _, _ int) ([]domainvalorant.Account, error) {
	return r.accounts, r.listErr
}

func testLogger() *logger.Logger {
	return &logger.Logger{SugaredLogger: zap.NewNop().Sugar()}
}

func testCfg() *config.Config {
	return &config.Config{
		Security: config.SecurityConfig{Secret: "test-secret"},
	}
}

func buildDeps(t *testing.T, userRepo user.Repository, valorantRepo domainvalorant.Repository) *deps.Deps {
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return &deps.Deps{
		Cfg:             testCfg(),
		Logging:         testLogger(),
		AuthTokenRepo:   redisstorage.NewAuthTokenRepo(rdb),
		PlayerNamesRepo: redisstorage.NewPlayerNamesRepo(rdb),
		AccountMetaRepo: redisstorage.NewAccountMetaRepo(rdb),
		UserSrv:         user.NewService(userRepo),
		ValorantSrv:     domainvalorant.NewService(valorantRepo),
	}
}

func newRouter(d *deps.Deps) *gin.Engine {
	r := gin.New()
	h := users.NewHandler(d)
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/logout", h.Logout)
	r.POST("/refresh", h.Refresh)
	r.GET("/me", func(c *gin.Context) { c.Set("user_id", int64(1)); h.Me(c) })
	r.GET("/accounts", func(c *gin.Context) { c.Set("user_id", int64(1)); h.GetAccounts(c) })
	return r
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func TestRegister_Success(t *testing.T) {
	stub := &stubUserRepo{stored: user.User{ID: 1, Login: "new_user"}}
	d := buildDeps(t, stub, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, map[string]string{
		"login": "new_user", "password": "pass123",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "new_user", resp["login"])

	cookies := w.Result().Cookies()
	var hasAccess, hasRefresh bool
	for _, c := range cookies {
		if c.Name == "access_token" {
			hasAccess = true
		}
		if c.Name == "refresh_token" {
			hasRefresh = true
		}
	}
	assert.True(t, hasAccess, "access_token cookie expected")
	assert.True(t, hasRefresh, "refresh_token cookie expected")
}

func TestRegister_InvalidBody(t *testing.T) {
	d := buildDeps(t, &stubUserRepo{}, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_DuplicateLogin(t *testing.T) {
	stub := &stubUserRepo{createErr: postgres.ErrLoginAlreadyTaken}
	d := buildDeps(t, stub, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/register", jsonBody(t, map[string]string{
		"login": "dup", "password": "pass",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success(t *testing.T) {
	hashed, err := hash.GeneratePasswordHash("pass")
	require.NoError(t, err)
	stub := &stubUserRepo{stored: user.User{ID: 2, Login: "user2", Password: hashed}}
	d := buildDeps(t, stub, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/login", jsonBody(t, map[string]string{
		"login": "user2", "password": "pass",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestLogin_InvalidBody(t *testing.T) {
	d := buildDeps(t, &stubUserRepo{}, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{bad json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	stub := &stubUserRepo{getErr: postgres.ErrUserNotFound}
	d := buildDeps(t, stub, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/login", jsonBody(t, map[string]string{
		"login": "nobody", "password": "pass",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	hashed, _ := hash.GeneratePasswordHash("correct")
	stub := &stubUserRepo{stored: user.User{ID: 1, Login: "u", Password: hashed}}
	d := buildDeps(t, stub, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/login", jsonBody(t, map[string]string{
		"login": "u", "password": "wrong",
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout_Success(t *testing.T) {
	d := buildDeps(t, &stubUserRepo{}, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRefresh_MissingCookie_Returns401(t *testing.T) {
	d := buildDeps(t, &stubUserRepo{}, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefresh_InvalidToken_Returns401(t *testing.T) {
	d := buildDeps(t, &stubUserRepo{}, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "no-such-token"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefresh_Success(t *testing.T) {
	d := buildDeps(t, &stubUserRepo{stored: user.User{ID: 5}}, &stubValorantRepo{})
	ctx := context.Background()

	require.NoError(t, d.AuthTokenRepo.SaveRefreshToken(ctx, "valid-uuid", 5))

	r := newRouter(d)
	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid-uuid"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	_, err := d.AuthTokenRepo.GetRefreshToken(ctx, "valid-uuid")
	assert.ErrorIs(t, err, redisstorage.ErrRefreshTokenNotFound)
}

func TestMe_Success(t *testing.T) {
	stub := &stubUserRepo{stored: user.User{ID: 1, Login: "me-user"}}
	d := buildDeps(t, stub, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, "me-user", resp["login"])
}

func TestMe_ServiceError_Returns500(t *testing.T) {
	stub := &stubUserRepo{getErr: context.DeadlineExceeded}
	d := buildDeps(t, stub, &stubValorantRepo{})
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetAccounts_EmptyList(t *testing.T) {
	stub := &stubValorantRepo{accounts: []domainvalorant.Account{}}
	d := buildDeps(t, &stubUserRepo{}, stub)
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestGetAccounts_ServiceError_Returns500(t *testing.T) {
	stub := &stubValorantRepo{listErr: context.DeadlineExceeded}
	d := buildDeps(t, &stubUserRepo{}, stub)
	r := newRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
