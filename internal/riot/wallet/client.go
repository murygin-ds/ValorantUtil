package wallet

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

func (c *Client) GetWallet(ctx context.Context) (*Wallet, error) {
	url := fmt.Sprintf("%s/store/v1/wallet/%s", c.base.PdURL(), c.base.PUUID())
	var raw rawWallet
	if err := c.base.Do(ctx, "GET", url, &raw); err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	return parse(raw), nil
}
