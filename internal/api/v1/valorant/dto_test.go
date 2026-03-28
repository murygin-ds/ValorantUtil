package valorant

import (
	"testing"
	"time"

	domainmatch "ValorantAPI/internal/domain/match"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToMatchDTO_MapsBasicFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	won := true
	m := domainmatch.Match{
		MatchID:           "match-1",
		MapID:             "map-1",
		QueueID:           "competitive",
		GameStartTime:     now,
		GameLengthMs:      1800000,
		TeamRedWon:        &won,
		TeamRedRoundsWon:  13,
		TeamBlueRoundsWon: 8,
	}

	dto := toMatchDTO(m)
	assert.Equal(t, "match-1", dto.MatchID)
	assert.Equal(t, "map-1", dto.MapID)
	assert.Equal(t, "competitive", dto.QueueID)
	assert.Equal(t, now, dto.PlayedAt)
	assert.Equal(t, int64(1800000), dto.DurationMs)
	require.NotNil(t, dto.TeamRedWon)
	assert.True(t, *dto.TeamRedWon)
	assert.Equal(t, 13, dto.TeamRedRoundsWon)
	assert.Equal(t, 8, dto.TeamBlueRoundsWon)
}

func TestToMatchDTO_PlayersAreSortedByScoreDescending(t *testing.T) {
	m := domainmatch.Match{
		Players: []domainmatch.Player{
			{PUUID: "p1", Score: 100},
			{PUUID: "p2", Score: 300},
			{PUUID: "p3", Score: 200},
		},
	}

	dto := toMatchDTO(m)
	require.Len(t, dto.Players, 3)
	assert.Equal(t, "p2", dto.Players[0].PUUID)
	assert.Equal(t, "p3", dto.Players[1].PUUID)
	assert.Equal(t, "p1", dto.Players[2].PUUID)
}

func TestToMatchDTO_KillsAreMapped(t *testing.T) {
	m := domainmatch.Match{
		MatchID: "m1",
		Kills: []domainmatch.Kill{
			{Round: 1, KillerPUUID: "killer", VictimPUUID: "victim", Assistants: []string{"assist"}},
		},
	}

	dto := toMatchDTO(m)
	require.Len(t, dto.Kills, 1)
	assert.Equal(t, 1, dto.Kills[0].Round)
	assert.Equal(t, "killer", dto.Kills[0].Killer)
	assert.Equal(t, "victim", dto.Kills[0].Victim)
	assert.Equal(t, []string{"assist"}, dto.Kills[0].Assistants)
}

func TestToMatchDTO_NilKillAssistantsBecomesEmptySlice(t *testing.T) {
	m := domainmatch.Match{
		Kills: []domainmatch.Kill{
			{KillerPUUID: "k", VictimPUUID: "v", Assistants: nil},
		},
	}
	dto := toMatchDTO(m)
	require.Len(t, dto.Kills, 1)
	assert.NotNil(t, dto.Kills[0].Assistants)
	assert.Empty(t, dto.Kills[0].Assistants)
}

func TestToMatchDTO_PlayerStatsMapped(t *testing.T) {
	m := domainmatch.Match{
		Players: []domainmatch.Player{
			{PUUID: "p1", TeamID: "Red", CharacterID: "agent-uuid", Score: 250, Kills: 20, Deaths: 5, Assists: 3},
		},
	}
	dto := toMatchDTO(m)
	require.Len(t, dto.Players, 1)
	p := dto.Players[0]
	assert.Equal(t, "p1", p.PUUID)
	assert.Equal(t, "Red", p.Team)
	assert.Equal(t, "agent-uuid", p.CharacterID)
	assert.Equal(t, 250, p.Score)
	assert.Equal(t, 20, p.Kills)
	assert.Equal(t, 5, p.Deaths)
	assert.Equal(t, 3, p.Assists)
}
