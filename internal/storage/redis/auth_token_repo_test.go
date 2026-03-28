package redis_test

import (
	"context"
	"testing"

	redisstorage "ValorantAPI/internal/storage/redis"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedis(t *testing.T) *goredis.Client {
	mr := miniredis.RunT(t)
	return goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
}

func TestGenerateRefreshTokenUUID_IsHex(t *testing.T) {
	uuid, err := redisstorage.GenerateRefreshTokenUUID()
	require.NoError(t, err)
	assert.Len(t, uuid, 32) // 16 bytes -> 32 hex chars
}

func TestGenerateRefreshTokenUUID_IsUnique(t *testing.T) {
	u1, _ := redisstorage.GenerateRefreshTokenUUID()
	u2, _ := redisstorage.GenerateRefreshTokenUUID()
	assert.NotEqual(t, u1, u2)
}

func TestAuthTokenRepo_SaveAndGet(t *testing.T) {
	repo := redisstorage.NewAuthTokenRepo(newTestRedis(t))
	ctx := context.Background()

	err := repo.SaveRefreshToken(ctx, "uuid-1", 42)
	require.NoError(t, err)

	userID, err := repo.GetRefreshToken(ctx, "uuid-1")
	require.NoError(t, err)
	assert.Equal(t, int64(42), userID)
}

func TestAuthTokenRepo_GetNonExistent_ReturnsNotFound(t *testing.T) {
	repo := redisstorage.NewAuthTokenRepo(newTestRedis(t))
	ctx := context.Background()

	_, err := repo.GetRefreshToken(ctx, "no-such-token")
	assert.ErrorIs(t, err, redisstorage.ErrRefreshTokenNotFound)
}

func TestAuthTokenRepo_Delete(t *testing.T) {
	repo := redisstorage.NewAuthTokenRepo(newTestRedis(t))
	ctx := context.Background()

	_ = repo.SaveRefreshToken(ctx, "uuid-del", 1)
	require.NoError(t, repo.DeleteRefreshToken(ctx, "uuid-del"))

	_, err := repo.GetRefreshToken(ctx, "uuid-del")
	assert.ErrorIs(t, err, redisstorage.ErrRefreshTokenNotFound)
}

func TestAuthTokenRepo_OverwriteToken(t *testing.T) {
	repo := redisstorage.NewAuthTokenRepo(newTestRedis(t))
	ctx := context.Background()

	_ = repo.SaveRefreshToken(ctx, "uuid-ow", 10)
	_ = repo.SaveRefreshToken(ctx, "uuid-ow", 20)

	userID, err := repo.GetRefreshToken(ctx, "uuid-ow")
	require.NoError(t, err)
	assert.Equal(t, int64(20), userID)
}
