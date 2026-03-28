package postgres

import (
	"ValorantAPI/internal/domain/valorant"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ValorantAccountRepo struct {
	pool *pgxpool.Pool
}

func NewValorantAccountRepo(pool *pgxpool.Pool) *ValorantAccountRepo {
	return &ValorantAccountRepo{pool: pool}
}

func (r *ValorantAccountRepo) CreateAccount(ctx context.Context, account *valorant.Account) error {
	query := `
        INSERT INTO valorant_accounts (user_id, puuid, region, shard)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (puuid) DO UPDATE SET
            user_id = EXCLUDED.user_id,
            region  = EXCLUDED.region,
            shard   = EXCLUDED.shard,
            updated_at = now()
        RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		account.UserID,
		account.PUUID,
		account.Region,
		account.Shard,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
}

func (r *ValorantAccountRepo) GetAccountsList(ctx context.Context, userID, limit, offset int) ([]valorant.Account, error) {
	query := `
		SELECT 
		    id, 
		    puuid, 
		    region,
		    shard,
		    updated_at,
		    created_at
		from valorant_accounts 
		where
		    user_id = $1
		limit $2 
		offset $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]valorant.Account, 0, limit)
	for rows.Next() {
		var account valorant.Account
		if err := rows.Scan(
			&account.ID,
			&account.PUUID,
			&account.Region,
			&account.Shard,
			&account.UpdatedAt,
			&account.CreatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}
