package redis_test

import (
	"context"
	"testing"

	redisstorage "ValorantAPI/internal/storage/redis"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlayerNamesRepo_GetAndSet(t *testing.T) {
	repo := redisstorage.NewPlayerNamesRepo(newTestRedis(t))
	ctx := context.Background()

	names := map[string]string{
		"puuid-1": "Player#1234",
		"puuid-2": "Hero#NA1",
	}
	require.NoError(t, repo.SetMany(ctx, names))

	for puuid, expected := range names {
		got, err := repo.Get(ctx, puuid)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	}
}

func TestPlayerNamesRepo_GetNonExistent_ReturnsNotFound(t *testing.T) {
	repo := redisstorage.NewPlayerNamesRepo(newTestRedis(t))
	ctx := context.Background()

	_, err := repo.Get(ctx, "ghost-puuid")
	assert.ErrorIs(t, err, redisstorage.ErrPlayerNameNotFound)
}

func TestPlayerNamesRepo_SetMany_EmptyMap_NoError(t *testing.T) {
	repo := redisstorage.NewPlayerNamesRepo(newTestRedis(t))
	ctx := context.Background()

	err := repo.SetMany(ctx, map[string]string{})
	assert.NoError(t, err)
}

func TestPlayerNamesRepo_SetMany_OverwritesExisting(t *testing.T) {
	repo := redisstorage.NewPlayerNamesRepo(newTestRedis(t))
	ctx := context.Background()

	_ = repo.SetMany(ctx, map[string]string{"p1": "Old#Name"})
	_ = repo.SetMany(ctx, map[string]string{"p1": "New#Name"})

	got, err := repo.Get(ctx, "p1")
	require.NoError(t, err)
	assert.Equal(t, "New#Name", got)
}
