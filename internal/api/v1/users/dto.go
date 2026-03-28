package users

import "ValorantAPI/internal/http/response"

type authUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authUserResponse struct {
	response.Response
	Login string `json:"login,omitempty"`
}

// GameName и TagLine заполняются из Redis (PlayerNamesRepo, TTL 7 дней).
// Tier и RR заполняются из Redis (AccountMetaRepo, TTL 1 час) после первого вызова /mmr.
type accountDTO struct {
	ID       int64  `json:"id"`
	PUUID    string `json:"puuid"`
	Region   string `json:"region"`
	Shard    string `json:"shard"`
	GameName string `json:"game_name,omitempty"`
	TagLine  string `json:"tag_line,omitempty"`
	Tier     int    `json:"tier"`
	RR       int    `json:"rr"`
}

type getAccountsResponse struct {
	response.Response
	Accounts []accountDTO `json:"accounts,omitempty"`
}

type meResponse struct {
	response.Response
	ID    int64  `json:"id,omitempty"`
	Login string `json:"login,omitempty"`
}
