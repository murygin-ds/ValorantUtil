package store

import (
	"ValorantAPI/internal/riot"
	"context"
	"fmt"
)

type Client struct {
	base *riot.Client
}

func NewClient(base *riot.Client) *Client {
	return &Client{base: base}
}

func (c *Client) GetStorefront(ctx context.Context) (*Storefront, error) {
	url := fmt.Sprintf("%s/store/v3/storefront/%s", c.base.PdURL(), c.base.PUUID())
	var result Storefront
	if err := c.base.Do(ctx, "POST", url, &result); err != nil {
		return nil, fmt.Errorf("get storefront: %w", err)
	}
	return &result, nil
}
