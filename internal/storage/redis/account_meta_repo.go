package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// accountMetaTTL - ранг может меняться после каждого матча, держим 1 час.
const accountMetaTTL = time.Hour

type AccountMeta struct {
	Tier int `json:"tier"`
	RR   int `json:"rr"`
}

type AccountMetaRepo struct {
	client *redis.Client
}

func NewAccountMetaRepo(client *redis.Client) *AccountMetaRepo {
	return &AccountMetaRepo{client: client}
}

func (r *AccountMetaRepo) Save(ctx context.Context, puuid string, meta AccountMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal account meta: %w", err)
	}
	if err := r.client.Set(ctx, accountMetaKey(puuid), data, accountMetaTTL).Err(); err != nil {
		return fmt.Errorf("save account meta cache: %w", err)
	}
	return nil
}

func (r *AccountMetaRepo) Get(ctx context.Context, puuid string) (AccountMeta, error) {
	val, err := r.client.Get(ctx, accountMetaKey(puuid)).Bytes()
	if errors.Is(err, redis.Nil) {
		return AccountMeta{}, ErrAccountMetaNotFound
	}
	if err != nil {
		return AccountMeta{}, fmt.Errorf("get account meta cache: %w", err)
	}
	var meta AccountMeta
	if err := json.Unmarshal(val, &meta); err != nil {
		return AccountMeta{}, fmt.Errorf("unmarshal account meta: %w", err)
	}
	return meta, nil
}

func accountMetaKey(puuid string) string { return "account_rank:" + puuid }
