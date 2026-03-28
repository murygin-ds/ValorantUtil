package match

// HistoryResponse - результат, полученный от эндпоинта competitive updates.
type HistoryResponse struct {
	Subject string
	History []HistoryItem
}

type HistoryItem struct {
	MatchID       string
	GameStartTime int64
	QueueID       string
	MapID         string
}

// competitiveUpdatesResponse - необработанный ответ от
// GET {pd}/mmr/v1/players/{puuid}/competitiveupdates
type competitiveUpdatesResponse struct {
	Version int                    `json:"Version"`
	Subject string                 `json:"Subject"`
	Matches []competitiveMatchItem `json:"Matches"`
}

type competitiveMatchItem struct {
	MatchID        string `json:"MatchID"`
	MapID          string `json:"MapID"`
	SeasonID       string `json:"SeasonID"`
	MatchStartTime int64  `json:"MatchStartTime"`
	QueueID        string `json:"QueueID"`
}

// MatchDetails возвращается запросом GET {pd}/match-details/v1/matches/{matchId}.
type MatchDetails struct {
	MatchInfo MatchInfo `json:"matchInfo"`
	Teams     []Team    `json:"teams"`
	Players   []Player  `json:"players"`
	Kills     []Kill    `json:"kills"`
}

type MatchInfo struct {
	MatchID          string `json:"matchId"`
	MapID            string `json:"mapId"`
	QueueID          string `json:"queueID"`
	GameLengthMillis int64  `json:"gameLength"`
	GameStartMillis  int64  `json:"gameStartMillis"`
}

type Team struct {
	TeamID       string `json:"teamId"`
	Won          bool   `json:"won"`
	RoundsWon    int    `json:"roundsWon"`
	RoundsPlayed int    `json:"roundsPlayed"`
}

type Player struct {
	Subject     string      `json:"subject"`
	TeamID      string      `json:"teamId"`
	CharacterID string      `json:"characterId"`
	Stats       PlayerStats `json:"stats"`
}

type PlayerStats struct {
	Score   int `json:"score"`
	Kills   int `json:"kills"`
	Deaths  int `json:"deaths"`
	Assists int `json:"assists"`
}

type Kill struct {
	Round      int      `json:"round"`
	Killer     string   `json:"killer"`
	Victim     string   `json:"victim"`
	Assistants []string `json:"assistants"`
}
