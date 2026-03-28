package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const matchesCacheTTL = 15 * time.Minute

type MatchesRepo struct {
	client *redis.Client
}

func NewMatchesRepo(client *redis.Client) *MatchesRepo {
	return &MatchesRepo{client: client}
}

// Save сохраняет JSON-кодированную историю матчей для игрока.
func (r *MatchesRepo) Save(ctx context.Context, puuid string, data []byte) error {
	if err := r.client.Set(ctx, matchesKey(puuid), data, matchesCacheTTL).Err(); err != nil {
		return fmt.Errorf("save matches cache: %w", err)
	}
	return nil
}

// Get возвращает кешированную историю матчей в JSON, или ErrMatchesNotFound при отсутствии/истечении.
func (r *MatchesRepo) Get(ctx context.Context, puuid string) ([]byte, error) {
	val, err := r.client.Get(ctx, matchesKey(puuid)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrMatchesNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get matches cache: %w", err)
	}
	return val, nil
}

// Invalidate удаляет кешированную историю матчей игрока (для принудительного обновления).
func (r *MatchesRepo) Invalidate(ctx context.Context, puuid string) error {
	return r.client.Del(ctx, matchesKey(puuid)).Err()
}

func matchesKey(puuid string) string { return "matches:" + puuid }
