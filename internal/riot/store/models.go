package store

type Storefront struct {
	FeaturedBundle       FeaturedBundle       `json:"FeaturedBundle"`
	SkinsPanelLayout     SkinsPanelLayout     `json:"SkinsPanelLayout"`
	UpgradeCurrencyStore UpgradeCurrencyStore `json:"UpgradeCurrencyStore"`
	AccessoryStore       AccessoryStore       `json:"AccessoryStore"`
	BonusStore           *BonusStore          `json:"BonusStore,omitempty"`
}

type Offer struct {
	OfferID          string         `json:"OfferID"`
	IsDirectPurchase bool           `json:"IsDirectPurchase"`
	StartDate        string         `json:"StartDate"`
	Cost             map[string]int `json:"Cost"`
	Rewards          []Reward       `json:"Rewards"`
}

type Reward struct {
	ItemTypeID string `json:"ItemTypeID"`
	ItemID     string `json:"ItemID"`
	Quantity   int    `json:"Quantity"`
}

// Бандлы

type FeaturedBundle struct {
	Bundle                           Bundle   `json:"Bundle"`
	Bundles                          []Bundle `json:"Bundles"`
	BundleRemainingDurationInSeconds int      `json:"BundleRemainingDurationInSeconds"`
}

type Bundle struct {
	ID                         string            `json:"ID"`
	DataAssetID                string            `json:"DataAssetID"`
	CurrencyID                 string            `json:"CurrencyID"`
	Items                      []BundleItem      `json:"Items"`
	ItemOffers                 []BundleItemOffer `json:"ItemOffers"`
	TotalBaseCost              map[string]int    `json:"TotalBaseCost"`
	TotalDiscountedCost        map[string]int    `json:"TotalDiscountedCost"`
	TotalDiscountPercent       float64           `json:"TotalDiscountPercent"`
	DurationRemainingInSeconds int               `json:"DurationRemainingInSeconds"`
	WholesaleOnly              bool              `json:"WholesaleOnly"`
}

type BundleItem struct {
	Item            BundleItemRef `json:"Item"`
	BasePrice       int           `json:"BasePrice"`
	CurrencyID      string        `json:"CurrencyID"`
	DiscountPercent float64       `json:"DiscountPercent"`
	DiscountedPrice int           `json:"DiscountedPrice"`
	IsPromoItem     bool          `json:"IsPromoItem"`
}

type BundleItemRef struct {
	ItemTypeID string `json:"ItemTypeID"`
	ItemID     string `json:"ItemID"`
	Amount     int    `json:"Amount"`
}

type BundleItemOffer struct {
	BundleItemOfferID string         `json:"BundleItemOfferID"`
	Offer             Offer          `json:"Offer"`
	DiscountPercent   float64        `json:"DiscountPercent"`
	DiscountedCost    map[string]int `json:"DiscountedCost"`
}

// Ежедневный магазин

type SkinsPanelLayout struct {
	SingleItemOffers                           []string `json:"SingleItemOffers"`
	SingleItemStoreOffers                      []Offer  `json:"SingleItemStoreOffers"`
	SingleItemOffersRemainingDurationInSeconds int      `json:"SingleItemOffersRemainingDurationInSeconds"`
}

// Валюта улучшений

type UpgradeCurrencyStore struct {
	UpgradeCurrencyOffers []UpgradeCurrencyOffer `json:"UpgradeCurrencyOffers"`
}

type UpgradeCurrencyOffer struct {
	OfferID           string  `json:"OfferID"`
	StorefrontItemID  string  `json:"StorefrontItemID"`
	Offer             Offer   `json:"Offer"`
	DiscountedPercent float64 `json:"DiscountedPercent"`
}

// Аксессуары

type AccessoryStore struct {
	AccessoryStoreOffers                     []AccessoryOffer `json:"AccessoryStoreOffers"`
	AccessoryStoreRemainingDurationInSeconds int              `json:"AccessoryStoreRemainingDurationInSeconds"`
	StorefrontID                             string           `json:"StorefrontID"`
}

type AccessoryOffer struct {
	Offer      Offer  `json:"Offer"`
	ContractID string `json:"ContractID"`
}

// Ночной рынок

type BonusStore struct {
	BonusStoreOffers                     []BonusOffer `json:"BonusStoreOffers"`
	BonusStoreRemainingDurationInSeconds int          `json:"BonusStoreRemainingDurationInSeconds"`
}

type BonusOffer struct {
	BonusOfferID    string         `json:"BonusOfferID"`
	Offer           Offer          `json:"Offer"`
	DiscountPercent float64        `json:"DiscountPercent"`
	DiscountCosts   map[string]int `json:"DiscountCosts"`
	IsSeen          bool           `json:"IsSeen"`
}
