package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// playerNameTTL - имена меняются редко, держим 7 дней.
const playerNameTTL = 7 * 24 * time.Hour

type PlayerNamesRepo struct {
	client *redis.Client
}

func NewPlayerNamesRepo(client *redis.Client) *PlayerNamesRepo {
	return &PlayerNamesRepo{client: client}
}

// Get возвращает кешированную строку "GameName#TagLine" для puuid.
// Возвращает ErrPlayerNameNotFound, если ключ отсутствует или истек.
func (r *PlayerNamesRepo) Get(ctx context.Context, puuid string) (string, error) {
	val, err := r.client.Get(ctx, playerNameKey(puuid)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrPlayerNameNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get player name: %w", err)
	}
	return val, nil
}

// SetMany сохраняет несколько записей puuid-"GameName#TagLine" в одном pipeline.
func (r *PlayerNamesRepo) SetMany(ctx context.Context, names map[string]string) error {
	if len(names) == 0 {
		return nil
	}
	pipe := r.client.Pipeline()
	for puuid, name := range names {
		pipe.Set(ctx, playerNameKey(puuid), name, playerNameTTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func playerNameKey(puuid string) string { return "player_name:" + puuid }
