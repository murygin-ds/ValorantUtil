package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const oauthStateTTL = 10 * time.Minute

type OAuthStateRepo struct {
	client *redis.Client
}

func NewOAuthStateRepo(client *redis.Client) *OAuthStateRepo {
	return &OAuthStateRepo{client: client}
}

// Save сохраняет маппинг state-строки и userID на время OAuth-потока.
func (r *OAuthStateRepo) Save(ctx context.Context, state string, userID int64) error {
	return r.client.Set(ctx, oauthStateKey(state), strconv.FormatInt(userID, 10), oauthStateTTL).Err()
}

// Consume получает userID для state и атомарно удаляет его.
// Возвращает ErrOAuthStateNotFound, если state неизвестен или истек.
func (r *OAuthStateRepo) Consume(ctx context.Context, state string) (int64, error) {
	key := oauthStateKey(state)

	val, err := r.client.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrOAuthStateNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get oauth state: %w", err)
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse user id: %w", err)
	}
	return userID, nil
}

func oauthStateKey(state string) string { return "oauth_state:" + state }
