package redis_test

import (
	"context"
	"testing"
	"time"

	redisstorage "ValorantAPI/internal/storage/redis"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorefrontRepo_SaveAndGet(t *testing.T) {
	repo := redisstorage.NewStorefrontRepo(newTestRedis(t))
	ctx := context.Background()

	data := []byte(`{"success":true}`)
	require.NoError(t, repo.Save(ctx, "puuid-1", data, time.Hour))

	got, err := repo.Get(ctx, "puuid-1")
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestStorefrontRepo_GetNonExistent_ReturnsNotFound(t *testing.T) {
	repo := redisstorage.NewStorefrontRepo(newTestRedis(t))
	ctx := context.Background()

	_, err := repo.Get(ctx, "no-puuid")
	assert.ErrorIs(t, err, redisstorage.ErrStorefrontNotFound)
}

func TestStorefrontRepo_Invalidate(t *testing.T) {
	repo := redisstorage.NewStorefrontRepo(newTestRedis(t))
	ctx := context.Background()

	_ = repo.Save(ctx, "puuid-inv", []byte(`{}`), time.Hour)
	require.NoError(t, repo.Invalidate(ctx, "puuid-inv"))

	_, err := repo.Get(ctx, "puuid-inv")
	assert.ErrorIs(t, err, redisstorage.ErrStorefrontNotFound)
}

func TestStorefrontRepo_TTLExpiry(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	repo := redisstorage.NewStorefrontRepo(rdb)
	ctx := context.Background()

	_ = repo.Save(ctx, "puuid-exp", []byte(`{}`), 5*time.Second)
	mr.FastForward(6 * time.Second)

	_, err := repo.Get(ctx, "puuid-exp")
	assert.ErrorIs(t, err, redisstorage.ErrStorefrontNotFound)
}
