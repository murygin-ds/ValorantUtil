package valorant

import (
	"ValorantAPI/internal/riot"
	"ValorantAPI/internal/riot/assets"
	"ValorantAPI/internal/riot/entitlements"
	"ValorantAPI/internal/riot/nameservice"
	"ValorantAPI/internal/riot/store"
	redisstorage "ValorantAPI/internal/storage/redis"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// storefrontTTL возвращает минимальную оставшуюся длительность среди всех разделов витрины.
// Ключ Redis истечет ровно в момент ротации первого раздела.
func storefrontTTL(sf *store.Storefront) time.Duration {
	candidates := []int{
		sf.SkinsPanelLayout.SingleItemOffersRemainingDurationInSeconds,
		sf.FeaturedBundle.BundleRemainingDurationInSeconds,
		sf.AccessoryStore.AccessoryStoreRemainingDurationInSeconds,
	}
	m := 0
	for _, s := range candidates {
		if s > 0 && (m == 0 || s < m) {
			m = s
		}
	}
	return time.Duration(m) * time.Second
}

// invalidateStorefrontCache удаляет кешированную витрину для puuid (например, при принудительном обновлении).
func (h *Handler) invalidateStorefrontCache(c *gin.Context, puuid string) {
	if err := h.deps.StorefrontRepo.Invalidate(c.Request.Context(), puuid); err != nil &&
		!errors.Is(err, redisstorage.ErrStorefrontNotFound) {
		h.deps.Logging.Warnw("failed to invalidate storefront cache", "puuid", puuid, "err", err)
	}
}

// enrichMatchPlayers получает имена игроков (Redis, TTL 7д) и данные агентов (PostgreSQL)
// для каждого игрока в срезе и заполняет поля Name/AgentName/AgentIcon на месте.
func (h *Handler) enrichMatchPlayers(c *gin.Context, riotClient *riot.Client, players []matchPlayerDTO) {
	ctx := c.Request.Context()

	// 1. Проверяем кеш; собираем PUUID, которые еще нужно загрузить.
	missing := make([]string, 0, len(players))
	cached := make(map[string]string, len(players))
	for _, p := range players {
		name, err := h.deps.PlayerNamesRepo.Get(ctx, p.PUUID)
		if err == nil {
			cached[p.PUUID] = name
		} else {
			missing = append(missing, p.PUUID)
		}
	}

	// 2. Массово загружаем недостающие имена из Riot name-service через уже созданный клиент.
	h.fillPlayerNames(c, riotClient, missing, cached)

	// 3. Заполняем поле Name каждого DTO игрока.
	for i := range players {
		if name, ok := cached[players[i].PUUID]; ok {
			players[i].Name = name
		}
	}

	for i := range players {
		if players[i].CharacterID == "" {
			continue
		}
		agentAsset, err := h.deps.AssetSrv.GetAsset(ctx, "agents", players[i].CharacterID)
		if err != nil {
			h.deps.Logging.Warnw("failed to resolve agent asset", "characterID", players[i].CharacterID, "err", err)
			continue
		}
		players[i].AgentName = agentAsset.DisplayNameEN
		players[i].AgentIcon = agentAsset.DisplayIconURL
	}
}

// fillPlayerNames получает недостающие имена игроков из Riot name-service и сохраняет их в Redis.
// Записывает результаты в переданный cached map для последующего поиска.
func (h *Handler) fillPlayerNames(c *gin.Context, riotClient *riot.Client, missing []string, cached map[string]string) {
	if len(missing) == 0 {
		return
	}
	ctx := c.Request.Context()

	names, err := nameservice.NewClient(riotClient).GetPlayerNames(ctx, missing)
	if err != nil {
		h.deps.Logging.Warnw("failed to fetch player names", "err", err)
		return
	}

	toStore := make(map[string]string, len(names))
	for _, n := range names {
		display := fmt.Sprintf("%s#%s", n.GameName, n.TagLine)
		cached[n.Subject] = display
		toStore[n.Subject] = display
	}
	if err := h.deps.PlayerNamesRepo.SetMany(ctx, toStore); err != nil {
		h.deps.Logging.Warnw("failed to cache player names", "err", err)
	}
}

// enrichOffers разрешает имя, цену и изображение для списка предложений магазина.
func (h *Handler) enrichOffers(c *gin.Context, offers []store.Offer) []storeItem {
	items := make([]storeItem, 0, len(offers))
	for _, offer := range offers {
		if len(offer.Rewards) == 0 {
			continue
		}
		reward := offer.Rewards[0]
		items = append(items, h.resolveItem(c, reward.ItemTypeID, reward.ItemID, offer.Cost))
	}
	return items
}

// enrichAccessoryOffers разрешает имя, цену и изображение для списка предложений аксессуаров.
func (h *Handler) enrichAccessoryOffers(c *gin.Context, offers []store.AccessoryOffer) []storeItem {
	items := make([]storeItem, 0, len(offers))
	for _, ao := range offers {
		if len(ao.Offer.Rewards) == 0 {
			continue
		}
		reward := ao.Offer.Rewards[0]
		items = append(items, h.resolveItem(c, reward.ItemTypeID, reward.ItemID, ao.Offer.Cost))
	}
	return items
}

// resolveItem получает данные ассета (имя + изображение) для одного предмета магазина с кешированием в БД.
func (h *Handler) resolveItem(c *gin.Context, typeUUID, itemID string, cost map[string]int) storeItem {
	item := storeItem{Price: firstCostValue(cost)}

	apiPath := assets.APIPathForTypeUUID(typeUUID)
	if apiPath == "" {
		return item
	}

	assetData, err := h.deps.AssetSrv.GetAsset(c.Request.Context(), apiPath, itemID)
	if err != nil {
		h.deps.Logging.Warnw("failed to resolve store item asset", "type", typeUUID, "item", itemID, "err", err)
		return item
	}

	item.Name = assetData.DisplayNameEN
	item.Image = assetData.DisplayIconURL
	return item
}

// enrichBundles разрешает имя, изображение и состав для всех рекомендуемых бандлов.
func (h *Handler) enrichBundles(c *gin.Context, featured store.FeaturedBundle) []bundleDTO {
	bundles := featured.Bundles
	// Запасной вариант: в некоторых ответах заполнено только единственное поле Bundle.
	if len(bundles) == 0 && featured.Bundle.DataAssetID != "" {
		bundles = []store.Bundle{featured.Bundle}
	}

	result := make([]bundleDTO, 0, len(bundles))
	for _, b := range bundles {
		dto := bundleDTO{
			TotalBasePrice:       firstCostValue(b.TotalBaseCost),
			TotalDiscountedPrice: firstCostValue(b.TotalDiscountedCost),
			TotalDiscountPercent: b.TotalDiscountPercent,
			ExpiresInSeconds:     b.DurationRemainingInSeconds,
		}

		// Получаем имя и изображение бандла через valorant-api.com/v1/bundles/{dataAssetID}
		if b.DataAssetID != "" {
			if bundleAsset, err := h.deps.AssetSrv.GetAsset(c.Request.Context(), assets.BundlesAPIPath, b.DataAssetID); err == nil {
				dto.Name = bundleAsset.DisplayNameEN
				dto.Image = bundleAsset.DisplayIconURL
			} else {
				h.deps.Logging.Warnw("failed to resolve bundle asset", "dataAssetID", b.DataAssetID, "err", err)
			}
		}

		// Разрешаем отдельные предметы бандла
		dto.Items = make([]bundleItem, 0, len(b.Items))
		for _, bi := range b.Items {
			item := bundleItem{
				BasePrice:       bi.BasePrice,
				DiscountedPrice: bi.DiscountedPrice,
				DiscountPercent: bi.DiscountPercent,
				IsPromo:         bi.IsPromoItem,
			}
			apiPath := assets.APIPathForTypeUUID(bi.Item.ItemTypeID)
			if apiPath != "" {
				if a, err := h.deps.AssetSrv.GetAsset(c.Request.Context(), apiPath, bi.Item.ItemID); err == nil {
					item.Name = a.DisplayNameEN
					item.Image = a.DisplayIconURL
				} else {
					h.deps.Logging.Warnw("failed to resolve bundle item asset", "type", bi.Item.ItemTypeID, "item", bi.Item.ItemID, "err", err)
				}
			}
			dto.Items = append(dto.Items, item)
		}

		result = append(result, dto)
	}
	return result
}

// firstCostValue возвращает первое значение из map стоимости (у предметов магазина обычно одна валюта).
func firstCostValue(cost map[string]int) int {
	for _, v := range cost {
		return v
	}
	return 0
}

// resolveSkins получает имеющиеся уровни скинов из Riot entitlements API и разрешает
// их имя и иконку через сервис ассетов (кеш PostgreSQL - valorant-api.com).
func (h *Handler) resolveSkins(c *gin.Context, riotClient *riot.Client) []skinDTO {
	ctx := c.Request.Context()

	owned, err := entitlements.NewClient(riotClient).GetByType(ctx, entitlements.TypeSkins)
	if err != nil {
		h.deps.Logging.Warnw("resolveSkins: failed to fetch skin entitlements", "err", err)
		return nil
	}
	if len(owned) == 0 {
		return nil
	}

	ids := make([]string, len(owned))
	for i, item := range owned {
		ids[i] = item.ItemID
	}

	apiPath := assets.APIPathForTypeUUID(entitlements.TypeSkins)
	assetsMap, err := h.deps.AssetSrv.BulkGetByType(ctx, entitlements.TypeSkins, apiPath, ids)
	if err != nil {
		h.deps.Logging.Warnw("resolveSkins: BulkGetByType failed", "err", err)
	}

	result := make([]skinDTO, 0, len(owned))
	for _, item := range owned {
		dto := skinDTO{ID: item.ItemID}
		if a, ok := assetsMap[item.ItemID]; ok {
			dto.Name = a.DisplayNameEN
			dto.Icon = a.DisplayIconURL
		}
		result = append(result, dto)
	}
	return result
}
