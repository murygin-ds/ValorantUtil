package riot

import (
	"ValorantAPI/internal/http/response"
	"net/url"
	"strings"
)

type riotCallbackRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
	IDToken     string `json:"id_token"`
}

type riotLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type riotLoginMFARequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Code      string `json:"code"       binding:"required"`
}

type riotLoginCaptchaRequest struct {
	SessionID    string `json:"session_id"    binding:"required"`
	Username     string `json:"username"      binding:"required"`
	Password     string `json:"password"      binding:"required"`
	CaptchaToken string `json:"captcha_token" binding:"required"`
}

type authURLResponse struct {
	response.Response
	AuthURL string `json:"auth_url,omitempty"`
}

// submitRedirectURLRequest отправляется после того, как пользователь скопирует URL из браузера,
// на который он попал после авторизации в Riot (https://playvalorant.com/opt_in#access_token=...).
type submitRedirectURLRequest struct {
	RedirectURL string `json:"redirect_url" binding:"required"`
}

type linkRiotAccountResponse struct {
	response.Response
	PUUID  string `json:"puuid,omitempty"`
	Region string `json:"region,omitempty"`
}

type riotLoginChallengeResponse struct {
	response.Response
	Status    string `json:"status"`
	SessionID string `json:"session_id"`
	// MFA
	Email      string `json:"email,omitempty"`
	Method     string `json:"method,omitempty"`
	CodeLength int    `json:"code_length,omitempty"`
	// Captcha
	HCaptchaKey string `json:"hcaptcha_key,omitempty"`
}

func buildRiotAuthURL(nonce string) string {
	u, _ := url.Parse("https://auth.riotgames.com/authorize")
	q := url.Values{}
	q.Set("client_id", "play-valorant-web-prod")
	q.Set("redirect_uri", "https://playvalorant.com/opt_in")
	q.Set("response_type", "token id_token")
	q.Set("scope", "account openid")
	q.Set("nonce", nonce)
	u.RawQuery = q.Encode()
	return u.String()
}

// parseRedirectURL извлекает access_token и id_token из URL-адреса перенаправления Riot.
// Riot помещает токены во фрагмент URL: https://playvalorant.com/opt_in#access_token=...&id_token=...
func parseRedirectURL(rawURL string) (accessToken, idToken string, err error) {
	normalized := strings.Replace(rawURL, "#", "?", 1)
	u, err := url.Parse(normalized)
	if err != nil {
		return "", "", err
	}
	accessToken = u.Query().Get("access_token")
	if accessToken == "" {
		return "", "", nil
	}
	idToken = u.Query().Get("id_token")
	return accessToken, idToken, nil
}
