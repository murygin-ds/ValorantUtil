package assets

type AssetData struct {
	UUID        string `json:"uuid"`
	DisplayName string `json:"displayName"`
	TitleText   string `json:"titleText"` // non-empty for player titles (in-game shown text)
	LevelItem   string `json:"levelItem"`
	DisplayIcon string `json:"displayIcon"`
	StreamVideo string `json:"streamedVideo"`
}

// Name returns the best display name: titleText (for titles) if non-empty, else displayName.
func (a AssetData) Name() string {
	if a.TitleText != "" {
		return a.TitleText
	}
	return a.DisplayName
}

type getAssetResponse struct {
	Status int       `json:"status"`
	Data   AssetData `json:"data"`
}

type getAllAssetsResponse struct {
	Status int         `json:"status"`
	Data   []AssetData `json:"data"`
}
