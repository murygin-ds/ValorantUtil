package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type AccountRepo struct {
	client *redis.Client
}

func NewAccountRepo(client *redis.Client) *AccountRepo {
	return &AccountRepo{client: client}
}

// Save сохраняет JSON-кодированный список скинов для puuid без срока истечения.
// Скины можно только добавлять к аккаунту, поэтому кеш остается актуальным всегда.
func (r *AccountRepo) Save(ctx context.Context, puuid string, data []byte) error {
	if err := r.client.Set(ctx, accountKey(puuid), data, 0).Err(); err != nil {
		return fmt.Errorf("save account cache: %w", err)
	}
	return nil
}

// Get возвращает кешированный JSON аккаунта для puuid, или ErrAccountNotFound при отсутствии/истечении.
func (r *AccountRepo) Get(ctx context.Context, puuid string) ([]byte, error) {
	val, err := r.client.Get(ctx, accountKey(puuid)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get account cache: %w", err)
	}
	return val, nil
}

// Invalidate удаляет кешированные данные аккаунта для puuid.
func (r *AccountRepo) Invalidate(ctx context.Context, puuid string) error {
	err := r.client.Del(ctx, accountKey(puuid)).Err()
	if errors.Is(err, redis.Nil) {
		return ErrAccountNotFound
	}
	return err
}

func accountKey(puuid string) string { return "account:" + puuid }
