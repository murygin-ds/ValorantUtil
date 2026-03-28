package assets

// itemTypePaths maps Riot's ItemTypeID UUIDs to valorant-api.com path segments.
// Used to resolve asset data (name, icon) for items returned by the Riot store API.
var itemTypePaths = map[string]string{
	"01bb38e1-da47-4e6a-9b3d-945fe4655707": "agents",
	"e7c63390-eda7-46e0-bb7a-a6abdacd2433": "weapons/skinlevels",
	"d5f120f8-ff8c-4aac-92ea-f2b5acbe9475": "sprays",
	"3f296c07-64c3-494c-923b-fe692a4fa1bd": "playertitles",
	"dd3bf334-87f3-40bd-b043-682a57a8dc3a": "buddies/levels",
	"2f09d728-3a17-6368-6f00-f80e5b04b2d6": "playercards",
}

// BundlesAPIPath is the valorant-api.com path for bundle data.
const BundlesAPIPath = "bundles"

// TitleTypeID is the Riot ItemTypeID UUID for player titles.
const TitleTypeID = "3f296c07-64c3-494c-923b-fe692a4fa1bd"

// APIPathForTypeUUID converts a Riot ItemTypeID UUID to the valorant-api.com path segment.
// Returns empty string if the UUID is not recognised.
func APIPathForTypeUUID(typeUUID string) string {
	return itemTypePaths[typeUUID]
}
