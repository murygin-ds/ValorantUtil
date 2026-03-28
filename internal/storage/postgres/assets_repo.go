package postgres

import (
	"ValorantAPI/internal/domain/asset"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AssetsRepo struct {
	pool *pgxpool.Pool
}

func NewAssetRepo(pool *pgxpool.Pool) *AssetsRepo {
	return &AssetsRepo{pool: pool}
}

func (r *AssetsRepo) GetAsset(ctx context.Context, itemTypeID, itemID string) (asset.Asset, error) {
	query := `
		SELECT 
		    id,
		    quantity,
		    price,
		    display_name_ru,
		    display_name_en,
		    display_icon_url,
		    stream_video_url,
		    created_at,
		    updated_at
		FROM
		    assets
		WHERE
		    type_id = $1 AND
		    item_id = $2
	`

	var itemAsset asset.Asset
	err := r.pool.QueryRow(ctx, query, itemTypeID, itemID).Scan(
		&itemAsset.ID,
		&itemAsset.Quantity,
		&itemAsset.Price,
		&itemAsset.DisplayNameRU,
		&itemAsset.DisplayNameEN,
		&itemAsset.DisplayIconURL,
		&itemAsset.StreamVideoURL,
		&itemAsset.CreatedAt,
		&itemAsset.UpdatedAt,
	)
	if err != nil {
		return asset.Asset{}, err
	}
	return itemAsset, nil
}

func (r *AssetsRepo) GetAssetsByTypeAndIDs(ctx context.Context, typeID string, ids []string) ([]asset.Asset, error) {
	query := `
		SELECT
		    id,
		    item_id,
		    quantity,
		    price,
		    display_name_ru,
		    display_name_en,
		    display_icon_url,
		    stream_video_url,
		    created_at,
		    updated_at
		FROM
		    assets
		WHERE
		    type_id = $1 AND
		    item_id = ANY($2)
	`

	rows, err := r.pool.Query(ctx, query, typeID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []asset.Asset
	for rows.Next() {
		var a asset.Asset
		a.TypeID = typeID
		if err := rows.Scan(
			&a.ID,
			&a.ItemID,
			&a.Quantity,
			&a.Price,
			&a.DisplayNameRU,
			&a.DisplayNameEN,
			&a.DisplayIconURL,
			&a.StreamVideoURL,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (r *AssetsRepo) AddAsset(ctx context.Context, itemAsset *asset.Asset) error {
	query := `
		INSERT INTO assets(
		    type_id,
		    item_id,
			quantity,
		    price,
		    display_name_ru,
		    display_name_en,
		    display_icon_url,
		    stream_video_url
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (type_id, item_id) DO UPDATE
		    SET display_name_en  = EXCLUDED.display_name_en,
		        display_icon_url = EXCLUDED.display_icon_url
		WHERE assets.display_name_en = '' AND assets.display_icon_url = ''
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		itemAsset.TypeID,
		itemAsset.ItemID,
		itemAsset.Quantity,
		itemAsset.Price,
		itemAsset.DisplayNameRU,
		itemAsset.DisplayNameEN,
		itemAsset.DisplayIconURL,
		itemAsset.StreamVideoURL,
	)
	return err
}
