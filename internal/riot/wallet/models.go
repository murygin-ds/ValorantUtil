package wallet

const (
	currencyVP             = "85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"
	currencyRadianite      = "e59aa87c-4cbf-517a-5983-6e81511be9b7"
	currencyKingdomCredits = "85ca954a-41f2-ce94-9b45-8ca3dd39a00d"
)

// rawWallet - структура, возвращаемая напрямую Riot API.
type rawWallet struct {
	Balances map[string]int `json:"Balances"`
}

// Wallet - разобранное, удобочитаемое представление баланса валюты игрока.
type Wallet struct {
	ValorantPoints  int `json:"valorant_points"`
	RadianitePoints int `json:"radianite_points"`
	KingdomCredits  int `json:"kingdom_credits"`
}

func parse(raw rawWallet) *Wallet {
	return &Wallet{
		ValorantPoints:  raw.Balances[currencyVP],
		RadianitePoints: raw.Balances[currencyRadianite],
		KingdomCredits:  raw.Balances[currencyKingdomCredits],
	}
}
