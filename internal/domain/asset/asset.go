package asset

import "time"

type Asset struct {
	ID int `json:"id"`

	TypeID   string `json:"type_id"`
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`

	DisplayNameRU  string `json:"display_name_ru"`
	DisplayNameEN  string `json:"display_name_en"`
	DisplayIconURL string `json:"display_icon_url"`
	StreamVideoURL string `json:"stream_video_url"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
