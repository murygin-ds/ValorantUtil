package entitlements

import (
	"ValorantAPI/internal/riot"
	"context"
	"encoding/json"
	"fmt"
)

const (
	TypeAgents  = "01bb38e1-da47-4e6a-9b3d-945fe4655707"
	TypeSkins   = "e7c63390-eda7-46e0-bb7a-a6abdacd2433"
	TypeSprays  = "d5f120f8-ff8c-4aac-92ea-f2b5acbe9475"
	TypeCards   = "2f09d728-3a17-6368-6f00-f80e5b04b2d6"
	TypeTitles  = "3f296c07-64c3-494c-923b-fe692a4fa1bd"
	TypeBuddies = "dd3bf334-87f3-40bd-b043-682a57a8dc3a"
)

// Item представляет один принадлежащий игроку предмет из API прав.
type Item struct {
	TypeID     string `json:"TypeID"`
	ItemID     string `json:"ItemID"`
	InstanceID string `json:"InstanceID"`
}

type entitlementsResponse struct {
	ItemTypeID   string `json:"ItemTypeID"`
	Entitlements []Item `json:"Entitlements"`
}

type Client struct {
	base *riot.Client
}

func NewClient(base *riot.Client) *Client {
	return &Client{base: base}
}

// GetByType возвращает все принадлежащие игроку предметы указанного itemTypeID.
func (c *Client) GetByType(ctx context.Context, itemTypeID string) ([]Item, error) {
	url := fmt.Sprintf("%s/store/v1/entitlements/%s/%s", c.base.PdURL(), c.base.PUUID(), itemTypeID)

	// Сначала декодируем в raw, чтобы при отладке видеть реальный ответ API.
	var raw json.RawMessage
	if err := c.base.Do(ctx, "GET", url, &raw); err != nil {
		return nil, fmt.Errorf("get entitlements (%s): %w", itemTypeID, err)
	}

	var resp entitlementsResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse entitlements (%s) body=%s: %w", itemTypeID, string(raw), err)
	}

	return resp.Entitlements, nil
}
