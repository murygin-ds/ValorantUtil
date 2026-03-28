package valorant

import "time"

type Account struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	PUUID     string    `json:"puuid"`
	Region    string    `json:"region"`
	Shard     string    `json:"shard"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
