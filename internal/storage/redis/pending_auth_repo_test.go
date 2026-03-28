package redis_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	redisstorage "ValorantAPI/internal/storage/redis"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPendingAuthRepo_SaveAndGet(t *testing.T) {
	repo := redisstorage.NewPendingAuthRepo(newTestRedis(t))
	ctx := context.Background()
	cookies := []*http.Cookie{{Name: "ssid", Value: "abc"}}

	require.NoError(t, repo.Save(ctx, "session-1", cookies))

	got, err := repo.Get(ctx, "session-1")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "ssid", got[0].Name)
	assert.Equal(t, "abc", got[0].Value)
}

func TestPendingAuthRepo_GetNonExistent_ReturnsNotFound(t *testing.T) {
	repo := redisstorage.NewPendingAuthRepo(newTestRedis(t))
	ctx := context.Background()

	_, err := repo.Get(ctx, "ghost")
	assert.ErrorIs(t, err, redisstorage.ErrPendingAuthNotFound)
}

func TestPendingAuthRepo_Delete(t *testing.T) {
	repo := redisstorage.NewPendingAuthRepo(newTestRedis(t))
	ctx := context.Background()
	cookies := []*http.Cookie{{Name: "c", Value: "v"}}

	_ = repo.Save(ctx, "session-del", cookies)
	require.NoError(t, repo.Delete(ctx, "session-del"))

	_, err := repo.Get(ctx, "session-del")
	assert.ErrorIs(t, err, redisstorage.ErrPendingAuthNotFound)
}

func TestPendingAuthRepo_ExpiredEntry_ReturnsNotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	repo := redisstorage.NewPendingAuthRepo(rdb)
	ctx := context.Background()
	cookies := []*http.Cookie{{Name: "c", Value: "v"}}

	require.NoError(t, repo.Save(ctx, "session-ttl", cookies))

	// Fast-forward miniredis clock past the TTL
	mr.FastForward(11 * time.Minute)

	_, err := repo.Get(ctx, "session-ttl")
	assert.ErrorIs(t, err, redisstorage.ErrPendingAuthNotFound)
}

func TestPendingAuthRepo_OverwriteCookies(t *testing.T) {
	repo := redisstorage.NewPendingAuthRepo(newTestRedis(t))
	ctx := context.Background()

	_ = repo.Save(ctx, "session-ow", []*http.Cookie{{Name: "old", Value: "v1"}})
	_ = repo.Save(ctx, "session-ow", []*http.Cookie{{Name: "new", Value: "v2"}})

	got, err := repo.Get(ctx, "session-ow")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "new", got[0].Name)
}
