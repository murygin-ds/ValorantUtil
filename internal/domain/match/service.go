package match

import (
	"ValorantAPI/internal/logger"
	riotmatch "ValorantAPI/internal/riot/match"
	"context"
	"time"
)

type Service struct {
	repo   Repository
	logger *logger.Logger
}

func NewService(repo Repository, logger *logger.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// SyncAndGet получает последние ID матчей игрока из Riot, сохраняет матчи,
// которых еще нет в БД, затем возвращает все сохраненные матчи игрока.
func (s *Service) SyncAndGet(ctx context.Context, riotClient *riotmatch.Client, puuid string) ([]Match, error) {
	history, err := riotClient.GetHistory(ctx, 0, 20)
	if err != nil {
		return nil, err
	}

	for _, item := range history.History {
		exists, err := s.repo.ExistsMatch(ctx, item.MatchID)
		if err != nil {
			s.logger.Warnw("failed to check match existence", "matchID", item.MatchID, "err", err)
			continue
		}
		if exists {
			continue
		}

		details, err := riotClient.GetDetails(ctx, item.MatchID)
		if err != nil {
			s.logger.Warnw("failed to fetch match details", "matchID", item.MatchID, "err", err)
			continue
		}

		m := fromRiotDetails(details)
		if err := s.repo.SaveMatch(ctx, m); err != nil {
			s.logger.Warnw("failed to save match", "matchID", item.MatchID, "err", err)
		}
	}

	return s.repo.GetMatchesByPUUID(ctx, puuid)
}

// GetCached возвращает сохраненные матчи из БД без вызова Riot API.
func (s *Service) GetCached(ctx context.Context, puuid string) ([]Match, error) {
	return s.repo.GetMatchesByPUUID(ctx, puuid)
}

// fromRiotDetails преобразует ответ Riot API о матче в доменную модель.
func fromRiotDetails(d *riotmatch.MatchDetails) *Match {
	m := &Match{
		MatchID:       d.MatchInfo.MatchID,
		MapID:         d.MatchInfo.MapID,
		QueueID:       d.MatchInfo.QueueID,
		GameStartTime: time.UnixMilli(d.MatchInfo.GameStartMillis),
		GameLengthMs:  d.MatchInfo.GameLengthMillis,
	}

	for _, t := range d.Teams {
		switch t.TeamID {
		case "Red":
			won := t.Won
			m.TeamRedWon = &won
			m.TeamRedRoundsWon = t.RoundsWon
		case "Blue":
			m.TeamBlueRoundsWon = t.RoundsWon
		}
	}

	for _, p := range d.Players {
		m.Players = append(m.Players, Player{
			MatchID:     d.MatchInfo.MatchID,
			PUUID:       p.Subject,
			TeamID:      p.TeamID,
			CharacterID: p.CharacterID,
			Score:       p.Stats.Score,
			Kills:       p.Stats.Kills,
			Deaths:      p.Stats.Deaths,
			Assists:     p.Stats.Assists,
		})
	}

	for _, k := range d.Kills {
		assistants := k.Assistants
		if assistants == nil {
			assistants = []string{}
		}
		m.Kills = append(m.Kills, Kill{
			MatchID:     d.MatchInfo.MatchID,
			Round:       k.Round,
			KillerPUUID: k.Killer,
			VictimPUUID: k.Victim,
			Assistants:  assistants,
		})
	}

	return m
}
