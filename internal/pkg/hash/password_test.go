package hash_test

import (
	"testing"

	"ValorantAPI/internal/pkg/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePasswordHash_Success(t *testing.T) {
	h, err := hash.GeneratePasswordHash("mypassword")
	require.NoError(t, err)
	assert.NotEmpty(t, h)
	assert.NotEqual(t, "mypassword", h)
}

func TestCheckPassword_Correct(t *testing.T) {
	h, err := hash.GeneratePasswordHash("correct")
	require.NoError(t, err)
	assert.True(t, hash.CheckPassword("correct", h))
}

func TestCheckPassword_Wrong(t *testing.T) {
	h, err := hash.GeneratePasswordHash("correct")
	require.NoError(t, err)
	assert.False(t, hash.CheckPassword("wrong", h))
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	h, err := hash.GeneratePasswordHash("nonempty")
	require.NoError(t, err)
	assert.False(t, hash.CheckPassword("", h))
}

func TestGeneratePasswordHash_SaltIsRandom(t *testing.T) {
	h1, err1 := hash.GeneratePasswordHash("same")
	h2, err2 := hash.GeneratePasswordHash("same")
	require.NoError(t, err1)
	require.NoError(t, err2)
	// bcrypt includes a random salt so identical passwords produce different hashes
	assert.NotEqual(t, h1, h2)
	// both hashes should still verify correctly
	assert.True(t, hash.CheckPassword("same", h1))
	assert.True(t, hash.CheckPassword("same", h2))
}
