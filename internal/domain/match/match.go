package match

import "time"

type Match struct {
	MatchID           string
	MapID             string
	QueueID           string
	GameStartTime     time.Time
	GameLengthMs      int64
	TeamRedWon        *bool
	TeamRedRoundsWon  int
	TeamBlueRoundsWon int
	Players           []Player
	Kills             []Kill
}

type Player struct {
	MatchID     string
	PUUID       string
	TeamID      string
	CharacterID string
	Score       int
	Kills       int
	Deaths      int
	Assists     int
}

type Kill struct {
	MatchID     string
	Round       int
	KillerPUUID string
	VictimPUUID string
	Assistants  []string
}
