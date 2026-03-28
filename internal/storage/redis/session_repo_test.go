package redis_test

import (
	"context"
	"net/http"
	"testing"

	"ValorantAPI/internal/riot/auth"
	redisstorage "ValorantAPI/internal/storage/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestSession() *auth.SessionData {
	return &auth.SessionData{
		AccessToken:      "access-token",
		IDToken:          "id-token",
		EntitlementToken: "entitlement-token",
		PUUID:            "test-puuid",
		Region:           "eu",
		Shard:            "eu",
		Cookies:          []*http.Cookie{{Name: "clid", Value: "test"}},
	}
}

func TestSessionRepo_SaveAndGet(t *testing.T) {
	repo := redisstorage.NewSessionRepo(newTestRedis(t))
	ctx := context.Background()

	session := makeTestSession()
	err := repo.SaveSession(ctx, session.PUUID, session)
	require.NoError(t, err)

	got, err := repo.GetSession(ctx, session.PUUID)
	require.NoError(t, err)
	assert.Equal(t, session.AccessToken, got.AccessToken)
	assert.Equal(t, session.EntitlementToken, got.EntitlementToken)
	assert.Equal(t, session.PUUID, got.PUUID)
	assert.Equal(t, session.Region, got.Region)
	assert.Equal(t, session.Shard, got.Shard)
}

func TestSessionRepo_GetNonExistent_ReturnsNotFound(t *testing.T) {
	repo := redisstorage.NewSessionRepo(newTestRedis(t))
	ctx := context.Background()

	_, err := repo.GetSession(ctx, "no-such-puuid")
	assert.ErrorIs(t, err, redisstorage.ErrSessionNotFound)
}

func TestSessionRepo_Delete(t *testing.T) {
	repo := redisstorage.NewSessionRepo(newTestRedis(t))
	ctx := context.Background()

	session := makeTestSession()
	_ = repo.SaveSession(ctx, session.PUUID, session)
	require.NoError(t, repo.DeleteSession(ctx, session.PUUID))

	_, err := repo.GetSession(ctx, session.PUUID)
	assert.ErrorIs(t, err, redisstorage.ErrSessionNotFound)
}

func TestSessionRepo_GetCookies(t *testing.T) {
	repo := redisstorage.NewSessionRepo(newTestRedis(t))
	ctx := context.Background()

	session := makeTestSession()
	_ = repo.SaveSession(ctx, session.PUUID, session)

	cookies, err := repo.GetCookies(ctx, session.PUUID)
	require.NoError(t, err)
	require.Len(t, cookies, 1)
	assert.Equal(t, "clid", cookies[0].Name)
	assert.Equal(t, "test", cookies[0].Value)
}

func TestSessionRepo_GetCookies_NotFound(t *testing.T) {
	repo := redisstorage.NewSessionRepo(newTestRedis(t))
	ctx := context.Background()

	_, err := repo.GetCookies(ctx, "ghost-puuid")
	assert.ErrorIs(t, err, redisstorage.ErrCookiesNotFound)
}
