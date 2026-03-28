package match

import "context"

type Repository interface {
	// ExistsMatch возвращает true, если матч уже сохранен.
	ExistsMatch(ctx context.Context, matchID string) (bool, error)
	// SaveMatch атомарно сохраняет полный матч (инфо + игроки + убийства).
	SaveMatch(ctx context.Context, m *Match) error
	// GetMatchesByPUUID возвращает все сохраненные матчи, в которых участвовал игрок,
	// от новых к старым.
	GetMatchesByPUUID(ctx context.Context, puuid string) ([]Match, error)
}
