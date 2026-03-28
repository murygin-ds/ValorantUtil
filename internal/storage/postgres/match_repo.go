package postgres

import (
	"ValorantAPI/internal/domain/match"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MatchRepo struct {
	pool *pgxpool.Pool
}

func NewMatchRepo(pool *pgxpool.Pool) *MatchRepo {
	return &MatchRepo{pool: pool}
}

func (r *MatchRepo) ExistsMatch(ctx context.Context, matchID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM matches WHERE match_id = $1)`,
		matchID,
	).Scan(&exists)
	return exists, err
}

func (r *MatchRepo) SaveMatch(ctx context.Context, m *match.Match) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO matches (
			match_id, map_id, queue_id, game_start_time, game_length_ms,
			team_red_won, team_red_rounds_won, team_blue_rounds_won
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (match_id) DO NOTHING`,
		m.MatchID, m.MapID, m.QueueID, m.GameStartTime, m.GameLengthMs,
		m.TeamRedWon, m.TeamRedRoundsWon, m.TeamBlueRoundsWon,
	)
	if err != nil {
		return err
	}

	for _, p := range m.Players {
		_, err = tx.Exec(ctx, `
			INSERT INTO match_players
				(match_id, puuid, team_id, character_id, score, kills, deaths, assists)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (match_id, puuid) DO NOTHING`,
			p.MatchID, p.PUUID, p.TeamID, p.CharacterID,
			p.Score, p.Kills, p.Deaths, p.Assists,
		)
		if err != nil {
			return err
		}
	}

	for _, k := range m.Kills {
		_, err = tx.Exec(ctx, `
			INSERT INTO match_kills (match_id, round, killer_puuid, victim_puuid, assistants)
			VALUES ($1,$2,$3,$4,$5)`,
			k.MatchID, k.Round, k.KillerPUUID, k.VictimPUUID, k.Assistants,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *MatchRepo) GetMatchesByPUUID(ctx context.Context, puuid string) ([]match.Match, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.match_id, m.map_id, m.queue_id, m.game_start_time, m.game_length_ms,
		       m.team_red_won, m.team_red_rounds_won, m.team_blue_rounds_won
		FROM matches m
		JOIN match_players mp ON mp.match_id = m.match_id
		WHERE mp.puuid = $1
		ORDER BY m.game_start_time DESC`,
		puuid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []match.Match
	for rows.Next() {
		var m match.Match
		if err := rows.Scan(
			&m.MatchID, &m.MapID, &m.QueueID, &m.GameStartTime, &m.GameLengthMs,
			&m.TeamRedWon, &m.TeamRedRoundsWon, &m.TeamBlueRoundsWon,
		); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Загружаем игроков и убийства для каждого матча.
	for i := range matches {
		players, err := r.getPlayers(ctx, matches[i].MatchID)
		if err != nil {
			return nil, err
		}
		matches[i].Players = players

		kills, err := r.getKills(ctx, matches[i].MatchID)
		if err != nil {
			return nil, err
		}
		matches[i].Kills = kills
	}

	return matches, nil
}

func (r *MatchRepo) getPlayers(ctx context.Context, matchID string) ([]match.Player, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT puuid, team_id, character_id, score, kills, deaths, assists
		FROM match_players WHERE match_id = $1`,
		matchID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []match.Player
	for rows.Next() {
		var p match.Player
		p.MatchID = matchID
		if err := rows.Scan(&p.PUUID, &p.TeamID, &p.CharacterID, &p.Score, &p.Kills, &p.Deaths, &p.Assists); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

func (r *MatchRepo) getKills(ctx context.Context, matchID string) ([]match.Kill, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT round, killer_puuid, victim_puuid, assistants
		FROM match_kills WHERE match_id = $1 ORDER BY round, id`,
		matchID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var kills []match.Kill
	for rows.Next() {
		var k match.Kill
		k.MatchID = matchID
		if err := rows.Scan(&k.Round, &k.KillerPUUID, &k.VictimPUUID, &k.Assistants); err != nil {
			return nil, err
		}
		kills = append(kills, k)
	}
	return kills, rows.Err()
}
