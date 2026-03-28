package jwt_test

import (
	"testing"
	"time"

	"ValorantAPI/internal/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "super-secret-key"

func TestGenerate_ParseRoundTrip(t *testing.T) {
	token, err := jwt.Generate(42, testSecret, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := jwt.Parse(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, int64(42), claims.UserID)
}

func TestParse_WrongSecret(t *testing.T) {
	token, err := jwt.Generate(1, testSecret, time.Hour)
	require.NoError(t, err)

	_, err = jwt.Parse(token, "wrong-secret")
	assert.Error(t, err)
}

func TestParse_ExpiredToken(t *testing.T) {
	token, err := jwt.Generate(1, testSecret, -time.Second)
	require.NoError(t, err)

	_, err = jwt.Parse(token, testSecret)
	assert.Error(t, err)
}

func TestParse_MalformedToken(t *testing.T) {
	_, err := jwt.Parse("not.a.jwt.token", testSecret)
	assert.Error(t, err)
}

func TestParse_EmptyToken(t *testing.T) {
	_, err := jwt.Parse("", testSecret)
	assert.Error(t, err)
}

func TestGenerate_DifferentUserIDs(t *testing.T) {
	t1, err1 := jwt.Generate(1, testSecret, time.Hour)
	t2, err2 := jwt.Generate(2, testSecret, time.Hour)
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, t1, t2)

	c1, _ := jwt.Parse(t1, testSecret)
	c2, _ := jwt.Parse(t2, testSecret)
	assert.Equal(t, int64(1), c1.UserID)
	assert.Equal(t, int64(2), c2.UserID)
}

func TestGenerate_ExpiryIsSet(t *testing.T) {
	duration := 15 * time.Minute
	before := time.Now()
	token, err := jwt.Generate(1, testSecret, duration)
	require.NoError(t, err)

	claims, err := jwt.Parse(token, testSecret)
	require.NoError(t, err)

	expiry := claims.ExpiresAt.Time
	assert.True(t, expiry.After(before.Add(14*time.Minute)))
	assert.True(t, expiry.Before(before.Add(16*time.Minute)))
}
