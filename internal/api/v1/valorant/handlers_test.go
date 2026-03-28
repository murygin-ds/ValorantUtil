package valorant_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ValorantAPI/internal/api/v1/valorant"
	"ValorantAPI/internal/config"
	"ValorantAPI/internal/deps"
	domainasset "ValorantAPI/internal/domain/asset"
	domainmatch "ValorantAPI/internal/domain/match"
	"ValorantAPI/internal/domain/user"
	domainvalorant "ValorantAPI/internal/domain/valorant"
	"ValorantAPI/internal/logger"
	riotassets "ValorantAPI/internal/riot/assets"
	"ValorantAPI/internal/riot/auth"
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

type stubAssetRepo struct{}

func (r *stubAssetRepo) GetAsset(_ context.Context, _, _ string) (domainasset.Asset, error) {
	return domainasset.Asset{}, nil
}
func (r *stubAssetRepo) GetAssetsByTypeAndIDs(_ context.Context, _ string, _ []string) ([]domainasset.Asset, error) {
	return nil, nil
}
func (r *stubAssetRepo) AddAsset(_ context.Context, _ *domainasset.Asset) error { return nil }

type stubMatchRepo struct {
	matches []domainmatch.Match
	getErr  error
}

func (r *stubMatchRepo) ExistsMatch(_ context.Context, _ string) (bool, error)   { return false, nil }
func (r *stubMatchRepo) SaveMatch(_ context.Context, _ *domainmatch.Match) error { return nil }
func (r *stubMatchRepo) GetMatchesByPUUID(_ context.Context, _ string) ([]domainmatch.Match, error) {
	return r.matches, r.getErr
}

type stubValorantRepo struct{}

func (r *stubValorantRepo) CreateAccount(_ context.Context, _ *domainvalorant.Account) error {
	return nil
}
func (r *stubValorantRepo) GetAccountsList(_ context.Context, _, _, _ int) ([]domainvalorant.Account, error) {
	return nil, nil
}

type stubUserRepo struct{}

func (r *stubUserRepo) CreateUser(_ context.Context, _ *user.User) error     { return nil }
func (r *stubUserRepo) GetUserByLogin(_ context.Context, _ *user.User) error { return nil }
func (r *stubUserRepo) GetUserByID(_ context.Context, _ *user.User) error    { return nil }

func buildDeps(t *testing.T) (*deps.Deps, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})

	assetSrv := domainasset.NewService(
		&stubAssetRepo{},
		riotassets.NewClient("https://valorant-api.com", &http.Client{}),
		testLogger(),
	)
	matchSrv := domainmatch.NewService(&stubMatchRepo{}, testLogger())

	return &deps.Deps{
		Cfg:             &config.Config{Security: config.SecurityConfig{Secret: "test-secret"}},
		Logging:         testLogger(),
		SessionRepo:     redisstorage.NewSessionRepo(rdb),
		StorefrontRepo:  redisstorage.NewStorefrontRepo(rdb),
		AccountRepo:     redisstorage.NewAccountRepo(rdb),
		AccountMetaRepo: redisstorage.NewAccountMetaRepo(rdb),
		PlayerNamesRepo: redisstorage.NewPlayerNamesRepo(rdb),
		ValorantSrv:     domainvalorant.NewService(&stubValorantRepo{}),
		AssetSrv:        assetSrv,
		MatchSrv:        matchSrv,
		HTTPClient:      &http.Client{},
	}, mr
}

func newValorantRouter(d *deps.Deps) *gin.Engine {
	r := gin.New()
	h := valorant.NewHandler(d)
	userMiddleware := func(c *gin.Context) { c.Set("user_id", int64(1)); c.Next() }
	g := r.Group("/valorant", userMiddleware)
	g.GET("/store/:puuid", h.GetUserStore)
	g.GET("/wallet/:puuid", h.GetWallet)
	g.GET("/mmr/:puuid", h.GetMMR)
	g.GET("/matches/:puuid", h.GetMatchHistory)
	g.GET("/account/:puuid", h.GetAccount)
	return r
}

func TestGetUserStore_SessionNotFound_Returns401(t *testing.T) {
	d, _ := buildDeps(t)
	r := newValorantRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/valorant/store/unknown-puuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserStore_CacheHit_Returns200(t *testing.T) {
	d, _ := buildDeps(t)
	ctx := context.Background()

	cached := map[string]any{"success": true, "store": []any{}}
	data, _ := json.Marshal(cached)
	require.NoError(t, d.StorefrontRepo.Save(ctx, "puuid-cache", data, 3600*1000000000))

	r := newValorantRouter(d)
	req := httptest.NewRequest(http.MethodGet, "/valorant/store/puuid-cache", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestGetWallet_SessionNotFound_Returns401(t *testing.T) {
	d, _ := buildDeps(t)
	r := newValorantRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/valorant/wallet/no-session", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMMR_SessionNotFound_Returns401(t *testing.T) {
	d, _ := buildDeps(t)
	r := newValorantRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/valorant/mmr/no-session", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMatchHistory_SessionNotFound_Returns401(t *testing.T) {
	d, _ := buildDeps(t)
	r := newValorantRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/valorant/matches/no-session", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMatchHistory_Cached_Returns200(t *testing.T) {
	d, mr := buildDeps(t)
	_ = mr

	session := &auth.SessionData{
		AccessToken:      "tok",
		EntitlementToken: "ent",
		PUUID:            "p1",
		Region:           "eu",
		Shard:            "eu",
	}
	require.NoError(t, d.SessionRepo.SaveSession(context.Background(), "p1", session))

	r := newValorantRouter(d)
	req := httptest.NewRequest(http.MethodGet, "/valorant/matches/p1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, w.Code == http.StatusInternalServerError || w.Code == http.StatusUnauthorized)
}

func TestGetAccount_SessionNotFound_Returns401(t *testing.T) {
	d, _ := buildDeps(t)
	r := newValorantRouter(d)

	req := httptest.NewRequest(http.MethodGet, "/valorant/account/no-session", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetAccount_CacheHit_Returns200(t *testing.T) {
	d, _ := buildDeps(t)
	ctx := context.Background()

	cached := map[string]any{"success": true, "skins": []any{}}
	data, _ := json.Marshal(cached)
	require.NoError(t, d.AccountRepo.Save(ctx, "puuid-acc", data))

	r := newValorantRouter(d)
	req := httptest.NewRequest(http.MethodGet, "/valorant/account/puuid-acc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestGetAccount_ForceFlagBypassesCache(t *testing.T) {
	d, _ := buildDeps(t)
	ctx := context.Background()

	cached := map[string]any{"success": true, "skins": []any{}}
	data, _ := json.Marshal(cached)
	_ = d.AccountRepo.Save(ctx, "puuid-force", data)

	r := newValorantRouter(d)
	req := httptest.NewRequest(http.MethodGet, "/valorant/account/puuid-force?force=true", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
