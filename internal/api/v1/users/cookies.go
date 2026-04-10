package users

import (
	"net/http"
	"time"
)

const (
	accessTokenDuration   = 15 * time.Minute
	accessTokenCookieAge  = 900    // 15 мин в секундах
	refreshTokenCookieAge = 604800 // 7 дней в секундах
)

func setAuthCookies(w http.ResponseWriter, accessToken, refreshUUID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   accessTokenCookieAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshUUID,
		Path:     "/",
		MaxAge:   refreshTokenCookieAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
