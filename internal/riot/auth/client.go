package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// ErrAccessTokenExpired возвращается, когда токен доступа Riot истек (HTTP 401).
var ErrAccessTokenExpired = errors.New("access token expired")

const riotAuthURL = "https://auth.riotgames.com"

type Client struct {
	httpClient *http.Client
	jar        http.CookieJar
}

func NewClient(httpClient *http.Client) *Client {
	jar, _ := cookiejar.New(nil)
	client := *httpClient
	client.Jar = jar
	return &Client{httpClient: &client, jar: jar}
}

// NewClientWithCookies создает клиент, который возобновляет существующую сессию авторизации,
// предварительно загружая сохраненные куки в jar.
func NewClientWithCookies(httpClient *http.Client, cookies []*http.Cookie) *Client {
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(riotAuthURL)
	jar.SetCookies(u, cookies)
	client := *httpClient
	client.Jar = jar
	return &Client{httpClient: &client, jar: jar}
}

// GetSessionCookies возвращает текущие куки сессии авторизации для сериализации.
func (c *Client) GetSessionCookies() []*http.Cookie {
	u, _ := url.Parse(riotAuthURL)
	return c.jar.Cookies(u)
}

func (c *Client) doJSON(ctx context.Context, method, rawURL string, body any, target any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RiotClient/99.0.0.0.0.0 rso-auth (Windows;10;;Professional, x64)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
	}
	return resp, nil
}

func (c *Client) InitAuthCookies(ctx context.Context) error {
	body := map[string]string{
		"client_id":     "play-valorant-web-prod",
		"nonce":         "1",
		"redirect_uri":  "https://playvalorant.com/opt_in",
		"response_type": "token id_token",
		"scope":         "account openid",
	}
	_, err := c.doJSON(ctx, http.MethodPost, riotAuthURL+"/api/v1/authorization", body, nil)
	return err
}

// submitAuthBody - внутренний метод, отправляющий тело авторизации и интерпретирующий ответ.
// Возвращает (redirectURI, challenge, error) - ровно одно из redirectURI или challenge будет не nil.
func (c *Client) submitAuthBody(ctx context.Context, body any) (redirectURI string, challenge *AuthChallenge, err error) {
	var result authRequestResponse
	if _, err := c.doJSON(ctx, http.MethodPut, riotAuthURL+"/api/v1/authorization", body, &result); err != nil {
		return "", nil, err
	}

	switch result.Type {
	case "success":
		if result.Success == nil {
			return "", nil, fmt.Errorf("success response missing parameters")
		}
		return result.Success.Parameters.URI, nil, nil

	case "multifactor":
		ch := &AuthChallenge{Type: ChallengeMFA}
		if result.Multifactor != nil {
			ch.Email = result.Multifactor.Email
			ch.Method = result.Multifactor.Method
			ch.CodeLength = result.Multifactor.MultiFactorCodeLength
		}
		return "", ch, nil

	case "auth":
		// Требуется капча; извлекаем ключ hcaptcha, если есть.
		ch := &AuthChallenge{Type: ChallengeCaptcha}
		if result.Captcha != nil && result.Captcha.HCaptcha != nil {
			ch.HCaptchaKey = result.Captcha.HCaptcha.Key
		}
		return "", ch, nil

	case "error", "":
		return "", nil, fmt.Errorf("riot auth error: %q", result.Error)

	default:
		return "", nil, fmt.Errorf("unexpected auth response type: %q", result.Type)
	}
}

// Authorize отправляет учетные данные (опционально с captchaToken вида "hcaptcha <token>").
// При успехе возвращает redirect URI. При challenge возвращает *ChallengeError.
func (c *Client) Authorize(ctx context.Context, username, password, captchaToken string) (string, error) {
	body := authRequestBody{
		Type:     "auth",
		Language: "en_US",
		Remember: true,
		RiotIdentity: riotIdentity{
			Captcha:  captchaToken,
			Username: username,
			Password: password,
		},
	}
	redirectURI, challenge, err := c.submitAuthBody(ctx, body)
	if err != nil {
		return "", err
	}
	if challenge != nil {
		return "", &ChallengeError{Challenge: *challenge}
	}
	return redirectURI, nil
}

// SubmitMFA отправляет код 2FA для сессии, которая уже вернула ChallengeMFA.
// При успехе возвращает redirect URI. Может вернуть *ChallengeError.
func (c *Client) SubmitMFA(ctx context.Context, code string) (string, error) {
	body := mfaRequestBody{
		Type:           "multifactor",
		Code:           code,
		RememberDevice: true,
	}
	redirectURI, challenge, err := c.submitAuthBody(ctx, body)
	if err != nil {
		return "", err
	}
	if challenge != nil {
		return "", &ChallengeError{Challenge: *challenge}
	}
	return redirectURI, nil
}

// CompleteAuth завершает процесс авторизации после получения успешного redirect URI.
func (c *Client) CompleteAuth(ctx context.Context, redirectURI string) (*SessionData, error) {
	accessToken, idToken, err := extractTokensFromURL(redirectURI)
	if err != nil {
		return nil, fmt.Errorf("extract tokens: %w", err)
	}

	entitlement, err := c.GetEntitlement(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("get entitlement: %w", err)
	}

	puuid, err := c.GetPlayerInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("get player info: %w", err)
	}

	region, shard, err := c.GetRiotGeo(ctx, accessToken, idToken)
	if err != nil {
		return nil, fmt.Errorf("get riot geo: %w", err)
	}

	u, _ := url.Parse(riotAuthURL)
	cookies := c.jar.Cookies(u)

	return &SessionData{
		AccessToken:      accessToken,
		IDToken:          idToken,
		EntitlementToken: entitlement,
		PUUID:            puuid,
		Region:           region,
		Shard:            shard,
		Cookies:          cookies,
	}, nil
}

func (c *Client) GetEntitlement(ctx context.Context, accessToken string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://entitlements.auth.riotgames.com/api/token/v1", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("%w: %s", ErrAccessTokenExpired, string(respBody))
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d with body %s", resp.StatusCode, string(respBody))
	}

	var result entitlementResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	return result.EntitlementsToken, nil
}

func (c *Client) GetPlayerInfo(ctx context.Context, accessToken string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		riotAuthURL+"/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result playerInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Sub, nil
}

func (c *Client) GetRiotGeo(ctx context.Context, accessToken, idToken string) (region, shard string, err error) {
	body := riotGeoBody{IDToken: idToken}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPut,
		"https://riot-geo.pas.si.riotgames.com/pas/v1/product/valorant", nil)

	data, _ := json.Marshal(body)
	req.Body = io.NopCloser(bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result riotGeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	return result.Affinities.Live, shardFromRegion(result.Affinities.Live), nil
}

// BuildSessionFromTokens создает SessionData из известных access_token и id_token,
// запрашивая entitlement, данные игрока и регион через Riot API.
// Используется, когда токены получены вне стандартного redirect-URI потока.
func (c *Client) BuildSessionFromTokens(ctx context.Context, accessToken, idToken string) (*SessionData, error) {
	entitlement, err := c.GetEntitlement(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("get entitlement: %w", err)
	}

	puuid, err := c.GetPlayerInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("get player info: %w", err)
	}

	region, shard, err := c.GetRiotGeo(ctx, accessToken, idToken)
	if err != nil {
		return nil, fmt.Errorf("get riot geo: %w", err)
	}

	return &SessionData{
		AccessToken:      accessToken,
		IDToken:          idToken,
		EntitlementToken: entitlement,
		PUUID:            puuid,
		Region:           region,
		Shard:            shard,
	}, nil
}

// Login выполняет полный процесс входа за один вызов.
// Возвращает ErrMFARequired или ErrCaptchaRequired, если требуется дополнительное взаимодействие.
func (c *Client) Login(ctx context.Context, username, password string) (*SessionData, error) {
	if err := c.InitAuthCookies(ctx); err != nil {
		return nil, fmt.Errorf("init cookies: %w", err)
	}

	redirectURI, err := c.Authorize(ctx, username, password, "")
	if err != nil {
		return nil, fmt.Errorf("authorize: %w", err)
	}

	return c.CompleteAuth(ctx, redirectURI)
}

func extractTokensFromURL(redirectURL string) (accessToken, idToken string, err error) {
	u, err := url.Parse(strings.Replace(redirectURL, "#", "?", 1))
	if err != nil {
		return "", "", err
	}
	accessToken = u.Query().Get("access_token")
	idToken = u.Query().Get("id_token")
	if accessToken == "" {
		return "", "", fmt.Errorf("access_token not found in redirect url")
	}
	return accessToken, idToken, nil
}

func shardFromRegion(region string) string {
	switch region {
	case "latam", "br", "na":
		return "na"
	case "pbe":
		return "pbe"
	default:
		return region
	}
}
