package mmr

type CompetitiveUpdate struct {
	MatchID                      string `json:"MatchID"`
	MapID                        string `json:"MapID"`
	SeasonID                     string `json:"SeasonID"`
	MatchStartTime               int64  `json:"MatchStartTime"`
	TierAfterUpdate              int    `json:"TierAfterUpdate"`
	TierBeforeUpdate             int    `json:"TierBeforeUpdate"`
	RankedRatingAfterUpdate      int    `json:"RankedRatingAfterUpdate"`
	RankedRatingBeforeUpdate     int    `json:"RankedRatingBeforeUpdate"`
	RankedRatingEarned           int    `json:"RankedRatingEarned"`
	RankedRatingPerformanceBonus int    `json:"RankedRatingPerformanceBonus"`
	TierProgressAfterUpdate      int    `json:"TierProgressAfterUpdate"`
	AFKPenalty                   int    `json:"AFKPenalty"`
}

type MMR struct {
	Version                     int               `json:"Version"`
	Subject                     string            `json:"Subject"`
	NewPlayerExperienceFinished bool              `json:"NewPlayerExperienceFinished"`
	QueueSkills                 map[string]any    `json:"QueueSkills"`
	LatestCompetitiveUpdate     CompetitiveUpdate `json:"LatestCompetitiveUpdate"`
	IsLeaderboardAnonymized     bool              `json:"IsLeaderboardAnonymized"`
	IsActRankBadgeHidden        bool              `json:"IsActRankBadgeHidden"`
}
