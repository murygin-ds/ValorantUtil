package riot

import (
	"ValorantAPI/internal/riot/auth"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	BaseURLPD     = "https://pd.%s.a.pvp.net"
	BaseURLGLZ    = "https://glz-%s-1.%s.a.pvp.net"
	BaseURLShared = "https://shared.%s.a.pvp.net"
)

type clientPlatform struct {
	PlatformType      string `json:"platformType"`
	PlatformOS        string `json:"platformOS"`
	PlatformOSVersion string `json:"platformOSVersion"`
	PlatformChipset   string `json:"platformChipset"`
}

var defaultPlatform = clientPlatform{
	PlatformType:      "PC",
	PlatformOS:        "Windows",
	PlatformOSVersion: "10.0.19042.1.256.64bit",
	PlatformChipset:   "Unknown",
}

func encodePlatform(p clientPlatform) string {
	data, _ := json.Marshal(p)
	return base64.StdEncoding.EncodeToString(data)
}

type Client struct {
	httpClient       *http.Client
	accessToken      string
	entitlementToken string
	puuid            string
	region           string
	shard            string
	clientVersion    string
}

type versionResponse struct {
	Data struct {
		RiotClientVersion string `json:"riotClientVersion"`
	} `json:"data"`
}

func fetchClientVersion(httpClient *http.Client) (string, error) {
	resp, err := httpClient.Get("https://valorant-api.com/v1/version")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result versionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.RiotClientVersion, nil
}

func NewClient(
	httpClient *http.Client,
	accessToken string,
	puuid string,
	region string,
	shard string,
) (*Client, error) {
	version, err := fetchClientVersion(httpClient)
	if err != nil {
		return nil, err
	}

	// Получаем entitlement-токен один раз при создании клиента, чтобы переиспользовать его
	// во всех последующих запросах без повторных обращений к авторизации.
	entitlement, err := auth.NewClient(httpClient).GetEntitlement(context.Background(), accessToken)
	if err != nil {
		return nil, fmt.Errorf("get entitlement token: %w", err)
	}

	return &Client{
		httpClient:       httpClient,
		accessToken:      accessToken,
		entitlementToken: entitlement,
		puuid:            puuid,
		region:           region,
		shard:            shard,
		clientVersion:    version,
	}, nil
}

func (c *Client) Do(ctx context.Context, method, url string, target any) error {
	return c.DoJSON(ctx, method, url, nil, target)
}

// DoJSON аналогичен Do, но сериализует body как JSON-нагрузку запроса.
// Для GET-запросов передайте nil body - тело не отправляется.
// Для POST/PUT/PATCH с nil body отправляется минимальный JSON-объект "{}".
func (c *Client) DoJSON(ctx context.Context, method, url string, body, target any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	} else if method != http.MethodGet && method != http.MethodHead && method != http.MethodDelete {
		// Запросы без явного body (не GET) отправляют минимальный JSON-объект.
		bodyReader = strings.NewReader("{}")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("X-Riot-Entitlements-JWT", c.entitlementToken)
	req.Header.Set("X-Riot-ClientVersion", c.clientVersion)
	req.Header.Set("X-Riot-ClientPlatform", encodePlatform(defaultPlatform))
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, body)
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) PdURL() string {
	return fmt.Sprintf(BaseURLPD, c.shard)
}

func (c *Client) GlzURL() string {
	return fmt.Sprintf(BaseURLGLZ, c.region, c.shard)
}

func (c *Client) SharedURL() string {
	return fmt.Sprintf(BaseURLShared, c.shard)
}

func (c *Client) PUUID() string            { return c.puuid }
func (c *Client) Region() string           { return c.region }
func (c *Client) Shard() string            { return c.shard }
func (c *Client) HTTPClient() *http.Client { return c.httpClient }
