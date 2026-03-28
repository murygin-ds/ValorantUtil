package mmr

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

func (c *Client) GetMMR(ctx context.Context) (*MMR, error) {
	url := fmt.Sprintf("%s/mmr/v1/players/%s", c.base.PdURL(), c.base.PUUID())
	var result MMR
	if err := c.base.Do(ctx, "GET", url, &result); err != nil {
		return nil, fmt.Errorf("get mmr: %w", err)
	}
	return &result, nil
}
