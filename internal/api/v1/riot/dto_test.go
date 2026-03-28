package riot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRedirectURL_WithBothTokens(t *testing.T) {
	rawURL := "https://playvalorant.com/opt_in#access_token=ACC&id_token=ID&token_type=Bearer&expires_in=3600"
	access, id, err := parseRedirectURL(rawURL)
	require.NoError(t, err)
	assert.Equal(t, "ACC", access)
	assert.Equal(t, "ID", id)
}

func TestParseRedirectURL_NoAccessToken(t *testing.T) {
	rawURL := "https://playvalorant.com/opt_in#id_token=ID"
	access, id, err := parseRedirectURL(rawURL)
	require.NoError(t, err)
	assert.Empty(t, access)
	assert.Empty(t, id)
}

func TestParseRedirectURL_InvalidURL(t *testing.T) {
	_, _, err := parseRedirectURL("://test")
	assert.Error(t, err)
}

func TestParseRedirectURL_PlainURL_NoFragment(t *testing.T) {
	rawURL := "https://playvalorant.com/opt_in"
	access, _, err := parseRedirectURL(rawURL)
	require.NoError(t, err)
	assert.Empty(t, access)
}

func TestBuildRiotAuthURL_ContainsRequiredParams(t *testing.T) {
	nonce := "test-nonce"
	u := buildRiotAuthURL(nonce)

	assert.Contains(t, u, "nonce="+nonce)
	assert.Contains(t, u, "client_id=play-valorant-web-prod")
	assert.Contains(t, u, "response_type=")
	assert.True(t, strings.HasPrefix(u, "https://auth.riotgames.com/authorize"))
}
