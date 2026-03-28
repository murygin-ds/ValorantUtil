package match

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

// GetHistory получает последние ID матчей через эндпоинт competitive updates,
// который является единственным эндпоинтом истории, работающим без регистрации разработчика.
// startIndex/endIndex управляют пагинацией (максимум 20 за вызов).
func (c *Client) GetHistory(ctx context.Context, startIndex, endIndex int) (*HistoryResponse, error) {
	url := fmt.Sprintf(
		"%s/mmr/v1/players/%s/competitiveupdates?queue=competitive&startIndex=%d&endIndex=%d",
		c.base.PdURL(), c.base.PUUID(), startIndex, endIndex,
	)

	var raw competitiveUpdatesResponse
	if err := c.base.Do(ctx, "GET", url, &raw); err != nil {
		return nil, fmt.Errorf("get match history: %w", err)
	}

	result := &HistoryResponse{Subject: raw.Subject}
	for _, m := range raw.Matches {
		queue := m.QueueID
		if queue == "" {
			queue = "competitive"
		}
		result.History = append(result.History, HistoryItem{
			MatchID:       m.MatchID,
			GameStartTime: m.MatchStartTime,
			QueueID:       queue,
			MapID:         m.MapID,
		})
	}
	return result, nil
}

// GetDetails возвращает полные данные матча по его ID.
func (c *Client) GetDetails(ctx context.Context, matchID string) (*MatchDetails, error) {
	url := fmt.Sprintf("%s/match-details/v1/matches/%s", c.base.PdURL(), matchID)
	var result MatchDetails
	if err := c.base.Do(ctx, "GET", url, &result); err != nil {
		return nil, fmt.Errorf("get match details %s: %w", matchID, err)
	}
	return &result, nil
}
