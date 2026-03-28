package valorant

import (
	"ValorantAPI/internal/domain/match"
	"ValorantAPI/internal/http/response"
	"ValorantAPI/internal/riot/wallet"
	"sort"
	"time"
)

type storeItem struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
	Image string `json:"image"`
}

type bundleItem struct {
	Name            string  `json:"name"`
	Image           string  `json:"image"`
	BasePrice       int     `json:"base_price"`
	DiscountedPrice int     `json:"discounted_price"`
	DiscountPercent float64 `json:"discount_percent"`
	IsPromo         bool    `json:"is_promo,omitempty"`
}

type bundleDTO struct {
	Name                 string       `json:"name"`
	Image                string       `json:"image"`
	Items                []bundleItem `json:"items"`
	TotalBasePrice       int          `json:"total_base_price"`
	TotalDiscountedPrice int          `json:"total_discounted_price"`
	TotalDiscountPercent float64      `json:"total_discount_percent"`
	ExpiresInSeconds     int          `json:"expires_in_seconds"`
}

type getDailyStoreResponse struct {
	response.Response
	Store       []storeItem `json:"store,omitempty"`
	Accessories []storeItem `json:"accessories,omitempty"`
	Bundles     []bundleDTO `json:"bundles,omitempty"`
}

type skinDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon,omitempty"`
}

type getAccountResponse struct {
	response.Response
	Skins []skinDTO `json:"skins,omitempty"`
}

type getWalletResponse struct {
	response.Response
	Wallet *wallet.Wallet `json:"wallet,omitempty"`
}

type mmrInfo struct {
	Tier         int    `json:"tier"`
	RankedRating int    `json:"ranked_rating"`
	LastMatchID  string `json:"last_match_id,omitempty"`
	LastMapID    string `json:"last_map_id,omitempty"`
	RRChange     int    `json:"rr_change"`
}

type getMMRResponse struct {
	response.Response
	MMR *mmrInfo `json:"mmr,omitempty"`
}

type matchKillDTO struct {
	Round      int      `json:"round"`
	Killer     string   `json:"killer"`
	Victim     string   `json:"victim"`
	Assistants []string `json:"assistants,omitempty"`
}

type matchPlayerDTO struct {
	PUUID       string `json:"puuid"`
	Name        string `json:"name"` // "GameName#TagLine"
	Team        string `json:"team"`
	CharacterID string `json:"character_id"`
	AgentName   string `json:"agent_name"`
	AgentIcon   string `json:"agent_icon"`
	Score       int    `json:"score"`
	Kills       int    `json:"kills"`
	Deaths      int    `json:"deaths"`
	Assists     int    `json:"assists"`
}

type matchDTO struct {
	MatchID           string           `json:"match_id"`
	MapID             string           `json:"map_id"`
	QueueID           string           `json:"queue_id"`
	PlayedAt          time.Time        `json:"played_at"`
	DurationMs        int64            `json:"duration_ms"`
	TeamRedWon        *bool            `json:"team_red_won"`
	TeamRedRoundsWon  int              `json:"team_red_rounds_won"`
	TeamBlueRoundsWon int              `json:"team_blue_rounds_won"`
	Players           []matchPlayerDTO `json:"players"`
	Kills             []matchKillDTO   `json:"kills,omitempty"`
}

type getMatchHistoryResponse struct {
	response.Response
	Matches []matchDTO `json:"matches,omitempty"`
}

func toMatchDTO(m match.Match) matchDTO {
	dto := matchDTO{
		MatchID:           m.MatchID,
		MapID:             m.MapID,
		QueueID:           m.QueueID,
		PlayedAt:          m.GameStartTime,
		DurationMs:        m.GameLengthMs,
		TeamRedWon:        m.TeamRedWon,
		TeamRedRoundsWon:  m.TeamRedRoundsWon,
		TeamBlueRoundsWon: m.TeamBlueRoundsWon,
		Players:           make([]matchPlayerDTO, 0, len(m.Players)),
		Kills:             make([]matchKillDTO, 0, len(m.Kills)),
	}
	for _, p := range m.Players {
		dto.Players = append(dto.Players, matchPlayerDTO{
			PUUID:       p.PUUID,
			Team:        p.TeamID,
			CharacterID: p.CharacterID,
			Score:       p.Score,
			Kills:       p.Kills,
			Deaths:      p.Deaths,
			Assists:     p.Assists,
		})
	}
	sort.Slice(dto.Players, func(i, j int) bool {
		return dto.Players[i].Score > dto.Players[j].Score
	})

	for _, k := range m.Kills {
		assistants := k.Assistants
		if assistants == nil {
			assistants = []string{}
		}
		dto.Kills = append(dto.Kills, matchKillDTO{
			Round:      k.Round,
			Killer:     k.KillerPUUID,
			Victim:     k.VictimPUUID,
			Assistants: assistants,
		})
	}
	return dto
}
