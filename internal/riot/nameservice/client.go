package nameservice

import (
	"ValorantAPI/internal/riot"
	"context"
	"fmt"
)

// PlayerName хранит отображаемое имя аккаунта Riot.
type PlayerName struct {
	Subject  string `json:"Subject"`  // puuid игрока
	GameName string `json:"GameName"` // например "PlayerOne"
	TagLine  string `json:"TagLine"`  // например "EUW"
}

type Client struct {
	base *riot.Client
}

func NewClient(base *riot.Client) *Client {
	return &Client{base: base}
}

// GetPlayerNames получает отображаемые имена для набора PUUID одним запросом.
// Использует PUT {pd}/name-service/v2/players с PUUID-листом в теле запроса.
func (c *Client) GetPlayerNames(ctx context.Context, puuids []string) ([]PlayerName, error) {
	url := fmt.Sprintf("%s/name-service/v2/players", c.base.PdURL())
	var result []PlayerName
	if err := c.base.DoJSON(ctx, "PUT", url, puuids, &result); err != nil {
		return nil, fmt.Errorf("get player names: %w", err)
	}
	return result, nil
}
