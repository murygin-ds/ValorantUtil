package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type StorefrontRepo struct {
	client *redis.Client
}

func NewStorefrontRepo(client *redis.Client) *StorefrontRepo {
	return &StorefrontRepo{client: client}
}

// Save сохраняет JSON-кодированную витрину для puuid с заданным TTL.
func (r *StorefrontRepo) Save(ctx context.Context, puuid string, data []byte, ttl time.Duration) error {
	if err := r.client.Set(ctx, storefrontKey(puuid), data, ttl).Err(); err != nil {
		return fmt.Errorf("save storefront cache: %w", err)
	}
	return nil
}

// Get возвращает кешированный JSON витрины для puuid, или ErrStorefrontNotFound при отсутствии/истечении.
func (r *StorefrontRepo) Get(ctx context.Context, puuid string) ([]byte, error) {
	val, err := r.client.Get(ctx, storefrontKey(puuid)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrStorefrontNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get storefront cache: %w", err)
	}
	return val, nil
}

// Invalidate удаляет кешированную витрину для puuid (для принудительного обновления).
func (r *StorefrontRepo) Invalidate(ctx context.Context, puuid string) error {
	err := r.client.Del(ctx, storefrontKey(puuid)).Err()
	if errors.Is(err, redis.Nil) {
		return ErrStorefrontNotFound
	}
	return err
}

func storefrontKey(puuid string) string { return "storefront:" + puuid }
