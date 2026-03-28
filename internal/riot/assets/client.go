package assets

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	AssetsAPIBaseURL string
	HTTPClient       *http.Client
}

func NewClient(assetsAPIBaseURL string, httpClient *http.Client) *Client {
	return &Client{
		AssetsAPIBaseURL: strings.TrimRight(assetsAPIBaseURL, "/"),
		HTTPClient:       httpClient,
	}
}

// GetAllByType получает все предметы указанного типа с valorant-api.com одним запросом.
func (c *Client) GetAllByType(itemType string) ([]AssetData, error) {
	req, err := http.NewRequest("GET", c.AssetsAPIBaseURL+"/v1/"+itemType, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d with body %s", resp.StatusCode, string(respBody))
	}

	var response getAllAssetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return response.Data, nil
}

func (c *Client) GetAsset(itemType, itemID string) (AssetData, error) {
	var data AssetData

	req, err := http.NewRequest("GET", c.AssetsAPIBaseURL+"/v1/"+itemType+"/"+itemID, nil)
	if err != nil {
		return data, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return data, fmt.Errorf("unexpected status code: %d with body %s", resp.StatusCode, string(respBody))
	}

	var response getAssetResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return data, fmt.Errorf("failed to decode response: %v", err)
	}

	return response.Data, nil
}
