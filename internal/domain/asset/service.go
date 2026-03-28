package asset

import (
	"ValorantAPI/internal/logger"
	"ValorantAPI/internal/riot/assets"
	"context"
	"database/sql"
	"errors"
)

type Service struct {
	client *assets.Client
	repo   Repository
	logger *logger.Logger
}

func NewService(repo Repository, client *assets.Client, logger *logger.Logger) *Service {
	return &Service{repo: repo, client: client, logger: logger}
}

func (s *Service) GetAsset(ctx context.Context, itemTypeID, itemID string) (Asset, error) {
	var asset Asset

	assetFromDB, err := s.repo.GetAsset(ctx, itemTypeID, itemID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return asset, err
		}
	} else {
		return assetFromDB, nil
	}

	assetFromAPI, err := s.client.GetAsset(itemTypeID, itemID)
	if err != nil {
		return asset, err
	}
	asset.TypeID = itemTypeID
	asset.ItemID = itemID
	asset.DisplayNameEN = assetFromAPI.DisplayName
	asset.DisplayIconURL = assetFromAPI.DisplayIcon

	go func() {
		if err := s.repo.AddAsset(context.Background(), &asset); err != nil {
			s.logger.Errorw("Failed to add asset to repository", "error", err)
		}
	}()

	return asset, nil
}

// BulkGetByType возвращает map[itemID]Asset для каждого id из ids.
// Найденные в БД предметы получает одним запросом; для отсутствующих в БД
// однократно вызывает API ассетов, чтобы получить полный список,
// и сохраняет недостающие предметы асинхронно.
func (s *Service) BulkGetByType(ctx context.Context, typeID, apiPath string, ids []string) (map[string]Asset, error) {
	result := make(map[string]Asset, len(ids))

	found, err := s.repo.GetAssetsByTypeAndIDs(ctx, typeID, ids)
	if err != nil {
		s.logger.Warnw("BulkGetByType: db lookup failed", "typeID", typeID, "err", err)
	}
	for _, a := range found {
		// Пропускаем устаревшие записи без имени и иконки - они будут повторно загружены.
		if a.DisplayNameEN != "" || a.DisplayIconURL != "" {
			result[a.ItemID] = a
		}
	}

	// Собираем ID, отсутствующие в БД (или устаревшие).
	var missing []string
	for _, id := range ids {
		if _, ok := result[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return result, nil
	}

	// Получаем полный список из внешнего API одним запросом.
	all, err := s.client.GetAllByType(apiPath)
	if err != nil {
		s.logger.Warnw("BulkGetByType: api fetch failed", "typeID", typeID, "apiPath", apiPath, "err", err)
		return result, nil // возвращаем что есть в БД
	}

	allByUUID := make(map[string]Asset, len(all))
	for _, a := range all {
		allByUUID[a.UUID] = Asset{
			TypeID:         typeID,
			ItemID:         a.UUID,
			DisplayNameEN:  a.Name(),
			DisplayIconURL: a.DisplayIcon,
		}
	}

	var toSave []*Asset
	for _, id := range missing {
		if a, ok := allByUUID[id]; ok {
			result[id] = a
			cp := a
			toSave = append(toSave, &cp)
		}
	}

	if len(toSave) > 0 {
		go func() {
			for _, a := range toSave {
				if err := s.repo.AddAsset(context.Background(), a); err != nil {
					s.logger.Errorw("BulkGetByType: failed to cache asset", "itemID", a.ItemID, "err", err)
				}
			}
		}()
	}

	return result, nil
}
