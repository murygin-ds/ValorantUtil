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

func TestAccountMetaRepo_SaveAndGet(t *testing.T) {
	repo := redisstorage.NewAccountMetaRepo(newTestRedis(t))
	ctx := context.Background()

	meta := redisstorage.AccountMeta{Tier: 21, RR: 75}
	require.NoError(t, repo.Save(ctx, "puuid-meta", meta))

	got, err := repo.Get(ctx, "puuid-meta")
	require.NoError(t, err)
	assert.Equal(t, meta.Tier, got.Tier)
	assert.Equal(t, meta.RR, got.RR)
}

func TestAccountMetaRepo_GetNonExistent_ReturnsNotFound(t *testing.T) {
	repo := redisstorage.NewAccountMetaRepo(newTestRedis(t))
	ctx := context.Background()

	_, err := repo.Get(ctx, "ghost")
	assert.ErrorIs(t, err, redisstorage.ErrAccountMetaNotFound)
}

func TestAccountMetaRepo_TTLExpiry(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	repo := redisstorage.NewAccountMetaRepo(rdb)
	ctx := context.Background()

	_ = repo.Save(ctx, "puuid-ttl", redisstorage.AccountMeta{Tier: 10, RR: 50})
	mr.FastForward(61 * time.Minute) // TTL is 1 hour

	_, err := repo.Get(ctx, "puuid-ttl")
	assert.ErrorIs(t, err, redisstorage.ErrAccountMetaNotFound)
}

func TestAccountMetaRepo_Overwrite(t *testing.T) {
	repo := redisstorage.NewAccountMetaRepo(newTestRedis(t))
	ctx := context.Background()

	_ = repo.Save(ctx, "p1", redisstorage.AccountMeta{Tier: 10, RR: 20})
	_ = repo.Save(ctx, "p1", redisstorage.AccountMeta{Tier: 15, RR: 99})

	got, err := repo.Get(ctx, "p1")
	require.NoError(t, err)
	assert.Equal(t, 15, got.Tier)
	assert.Equal(t, 99, got.RR)
}
