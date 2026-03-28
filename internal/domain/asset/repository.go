package asset

import "context"

type Repository interface {
	GetAsset(ctx context.Context, itemTypeID, itemID string) (Asset, error)
	GetAssetsByTypeAndIDs(ctx context.Context, typeID string, ids []string) ([]Asset, error)
	AddAsset(ctx context.Context, itemAsset *Asset) error
}
