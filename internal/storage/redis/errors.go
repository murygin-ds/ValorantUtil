package redis

import "fmt"

var (
	ErrSessionNotFound      = fmt.Errorf("session not found")
	ErrCookiesNotFound      = fmt.Errorf("cookies not found")
	ErrRefreshTokenNotFound = fmt.Errorf("refresh token not found")
	ErrPendingAuthNotFound  = fmt.Errorf("pending auth session not found or expired")
	ErrOAuthStateNotFound   = fmt.Errorf("oauth state not found or expired")
	ErrStorefrontNotFound   = fmt.Errorf("storefront cache not found or expired")
	ErrMatchesNotFound      = fmt.Errorf("matches cache not found or expired")
	ErrPlayerNameNotFound   = fmt.Errorf("player name not found in cache")
	ErrAccountNotFound      = fmt.Errorf("account cache not found or expired")
	ErrAccountMetaNotFound  = fmt.Errorf("account meta cache not found or expired")
)
