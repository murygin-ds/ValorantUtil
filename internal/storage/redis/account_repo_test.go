package redis_test

import (
	"context"
	"testing"

	redisstorage "ValorantAPI/internal/storage/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountRepo_SaveAndGet(t *testing.T) {
	repo := redisstorage.NewAccountRepo(newTestRedis(t))
	ctx := context.Background()

	data := []byte(`{"success":true,"skins":[]}`)
	require.NoError(t, repo.Save(ctx, "puuid-acc", data))

	got, err := repo.Get(ctx, "puuid-acc")
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestAccountRepo_GetNonExistent_ReturnsNotFound(t *testing.T) {
	repo := redisstorage.NewAccountRepo(newTestRedis(t))
	ctx := context.Background()

	_, err := repo.Get(ctx, "ghost-puuid")
	assert.ErrorIs(t, err, redisstorage.ErrAccountNotFound)
}

func TestAccountRepo_Invalidate(t *testing.T) {
	repo := redisstorage.NewAccountRepo(newTestRedis(t))
	ctx := context.Background()

	_ = repo.Save(ctx, "puuid-inv", []byte(`{}`))
	require.NoError(t, repo.Invalidate(ctx, "puuid-inv"))

	_, err := repo.Get(ctx, "puuid-inv")
	assert.ErrorIs(t, err, redisstorage.ErrAccountNotFound)
}

func TestAccountRepo_NoExpiry(t *testing.T) {
	// Account repo stores skins permanently - no TTL should be set
	repo := redisstorage.NewAccountRepo(newTestRedis(t))
	ctx := context.Background()

	data := []byte(`{"skins":["s1","s2"]}`)
	require.NoError(t, repo.Save(ctx, "puuid-perm", data))

	// Data should still be there after retrieval without any TTL manipulation
	got, err := repo.Get(ctx, "puuid-perm")
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestAccountRepo_Overwrite(t *testing.T) {
	repo := redisstorage.NewAccountRepo(newTestRedis(t))
	ctx := context.Background()

	_ = repo.Save(ctx, "p1", []byte(`{"v":1}`))
	_ = repo.Save(ctx, "p1", []byte(`{"v":2}`))

	got, err := repo.Get(ctx, "p1")
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"v":2}`), got)
}
