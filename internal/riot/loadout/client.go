package loadout

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

type identity struct {
	PlayerCardID  string `json:"PlayerCardID"`
	PlayerTitleID string `json:"PlayerTitleID"`
	AccountLevel  int    `json:"AccountLevel"`
}

type PlayerLoadout struct {
	Subject  string   `json:"Subject"`
	Identity identity `json:"Identity"`
}

// GetPlayerLoadout возвращает текущий экипированный набор предметов игрока.
func (c *Client) GetPlayerLoadout(ctx context.Context) (*PlayerLoadout, error) {
	url := fmt.Sprintf("%s/personalization/v2/players/%s/playerloadout", c.base.PdURL(), c.base.PUUID())
	var result PlayerLoadout
	if err := c.base.Do(ctx, "GET", url, &result); err != nil {
		return nil, fmt.Errorf("get player loadout: %w", err)
	}
	return &result, nil
}
