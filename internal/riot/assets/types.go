package assets

// itemTypePaths maps Riot's ItemTypeID UUIDs to valorant-api.com path segments.
// Used to resolve asset data (name, icon) for items returned by the Riot store API.
var itemTypePaths = map[string]string{
	"01bb38e1-da47-4e6a-9b3d-945fe4655707": "agents",
	"e7c63390-eda7-46e0-bb7a-a6abdacd2433": "weapons/skinlevels",
	"d5f120f8-ff8c-4aac-92ea-f2b5acbe9475": "sprays",
	"de7caa6b-adf7-4588-bbd1-143831e786c6": "playertitles",
	"dd3bf334-87f3-40bd-b043-682a57a8dc3a": "buddies/levels",
	"3f296c07-64c3-494c-923b-fe692a4fa1bd": "playercards",
}

// BundlesAPIPath is the valorant-api.com path for bundle data.
const BundlesAPIPath = "bundles"

// TitleTypeID is the Riot ItemTypeID UUID for player titles.
const TitleTypeID = "de7caa6b-adf7-4588-bbd1-143831e786c6"

// APIPathForTypeUUID converts a Riot ItemTypeID UUID to the valorant-api.com path segment.
// Returns empty string if the UUID is not recognised.
func APIPathForTypeUUID(typeUUID string) string {
	return itemTypePaths[typeUUID]
}
