package content

import (
	"ValorantAPI/internal/riot"
	"ValorantAPI/internal/riot/assets"
	"context"
	"encoding/json"
	"fmt"
)

type Client struct {
	base *riot.Client
}

func NewClient(base *riot.Client) *Client {
	return &Client{base: base}
}

type offerReward struct {
	ItemTypeID string `json:"ItemTypeID"`
	ItemID     string `json:"ItemID"`
}

type offer struct {
	OfferID string        `json:"OfferID"`
	Rewards []offerReward `json:"Rewards"`
}

type offersResponse struct {
	Offers []offer `json:"Offers"`
}

// TitleOfferMap получает все предложения магазина и возвращает map[offerID]contentUUID
// для предложений, наградой которых является титул игрока.
// Полученный contentUUID используется для поиска названий титулов на valorant-api.com.
func (c *Client) TitleOfferMap(ctx context.Context) (map[string]string, error) {
	url := fmt.Sprintf("%s/store/v1/offers", c.base.PdURL())

	var raw json.RawMessage
	if err := c.base.Do(ctx, "GET", url, &raw); err != nil {
		return nil, fmt.Errorf("get offers (url=%s): %w", url, err)
	}

	var resp offersResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse offers: %w", err)
	}

	out := make(map[string]string)
	for _, o := range resp.Offers {
		if len(o.Rewards) > 0 && o.Rewards[0].ItemTypeID == assets.TitleTypeID {
			out[o.OfferID] = o.Rewards[0].ItemID
		}
	}
	return out, nil
}
