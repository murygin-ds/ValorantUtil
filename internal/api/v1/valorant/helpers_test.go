package valorant

import (
	"testing"
	"time"

	"ValorantAPI/internal/riot/store"

	"github.com/stretchr/testify/assert"
)

func TestFirstCostValue_NonEmpty(t *testing.T) {
	cost := map[string]int{"currency-uuid": 1775}
	assert.Equal(t, 1775, firstCostValue(cost))
}

func TestFirstCostValue_Empty(t *testing.T) {
	assert.Equal(t, 0, firstCostValue(nil))
	assert.Equal(t, 0, firstCostValue(map[string]int{}))
}

func TestStorefrontTTL_UsesMinimumNonZero(t *testing.T) {
	sf := &store.Storefront{}
	sf.SkinsPanelLayout.SingleItemOffersRemainingDurationInSeconds = 3600
	sf.FeaturedBundle.BundleRemainingDurationInSeconds = 1800
	sf.AccessoryStore.AccessoryStoreRemainingDurationInSeconds = 7200

	ttl := storefrontTTL(sf)
	assert.Equal(t, 1800*time.Second, ttl)
}

func TestStorefrontTTL_AllZero(t *testing.T) {
	sf := &store.Storefront{}
	assert.Equal(t, time.Duration(0), storefrontTTL(sf))
}

func TestStorefrontTTL_OneNonZero(t *testing.T) {
	sf := &store.Storefront{}
	sf.SkinsPanelLayout.SingleItemOffersRemainingDurationInSeconds = 500
	assert.Equal(t, 500*time.Second, storefrontTTL(sf))
}

func TestStorefrontTTL_IgnoresZeroValues(t *testing.T) {
	sf := &store.Storefront{}
	sf.SkinsPanelLayout.SingleItemOffersRemainingDurationInSeconds = 0
	sf.FeaturedBundle.BundleRemainingDurationInSeconds = 900
	sf.AccessoryStore.AccessoryStoreRemainingDurationInSeconds = 0

	ttl := storefrontTTL(sf)
	assert.Equal(t, 900*time.Second, ttl)
}
