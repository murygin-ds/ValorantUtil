package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const pendingAuthTTL = 10 * time.Minute

type pendingAuthData struct {
	Cookies []*http.Cookie `json:"cookies"`
}

type PendingAuthRepo struct {
	client *redis.Client
}

func NewPendingAuthRepo(client *redis.Client) *PendingAuthRepo {
	return &PendingAuthRepo{client: client}
}

// Save сохраняет состояние cookie-jar для текущей сессии авторизации Riot.
// session_id - краткосрочный ключ для связи последующих запросов (MFA, капча).
func (r *PendingAuthRepo) Save(ctx context.Context, sessionID string, cookies []*http.Cookie) error {
	data := pendingAuthData{Cookies: cookies}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal pending auth: %w", err)
	}
	return r.client.Set(ctx, pendingAuthKey(sessionID), b, pendingAuthTTL).Err()
}

// Get возвращает куки для session_id, или ErrPendingAuthNotFound если ключ истек.
func (r *PendingAuthRepo) Get(ctx context.Context, sessionID string) ([]*http.Cookie, error) {
	val, err := r.client.Get(ctx, pendingAuthKey(sessionID)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrPendingAuthNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get pending auth: %w", err)
	}

	var data pendingAuthData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal pending auth: %w", err)
	}
	return data.Cookies, nil
}

// Delete удаляет состояние ожидающей авторизации (например, при успехе или превышении попыток).
func (r *PendingAuthRepo) Delete(ctx context.Context, sessionID string) error {
	return r.client.Del(ctx, pendingAuthKey(sessionID)).Err()
}

func pendingAuthKey(sessionID string) string { return "riot_login_pending:" + sessionID }
