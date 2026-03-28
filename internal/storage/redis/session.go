package redis

import (
	"ValorantAPI/internal/riot/auth"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionTTL = 20 * 24 * time.Hour
	cookiesTTL = 20 * 24 * time.Hour
)

type sessionData struct {
	AccessToken      string `json:"access_token"`
	IDToken          string `json:"id_token"`
	EntitlementToken string `json:"entitlement_token"`
	PUUID            string `json:"puuid"`
	Region           string `json:"region"`
	Shard            string `json:"shard"`
}

type cookieData struct {
	Cookies []*http.Cookie `json:"cookies"`
}

type SessionRepo struct {
	client *redis.Client
}

func NewSessionRepo(client *redis.Client) *SessionRepo {
	return &SessionRepo{client: client}
}

func (r *SessionRepo) SaveSession(ctx context.Context, puuid string, session *auth.SessionData) error {
	data := sessionData{
		AccessToken:      session.AccessToken,
		IDToken:          session.IDToken,
		EntitlementToken: session.EntitlementToken,
		PUUID:            session.PUUID,
		Region:           session.Region,
		Shard:            session.Shard,
	}

	sessionJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	cookiesJSON, err := json.Marshal(cookieData{Cookies: session.Cookies})
	if err != nil {
		return fmt.Errorf("marshal cookies: %w", err)
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, sessionKey(puuid), sessionJSON, sessionTTL)
	pipe.Set(ctx, cookiesKey(puuid), cookiesJSON, cookiesTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *SessionRepo) GetSession(ctx context.Context, puuid string) (*auth.SessionData, error) {
	val, err := r.client.Get(ctx, sessionKey(puuid)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	var data sessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}

	return &auth.SessionData{
		AccessToken:      data.AccessToken,
		IDToken:          data.IDToken,
		EntitlementToken: data.EntitlementToken,
		PUUID:            data.PUUID,
		Region:           data.Region,
		Shard:            data.Shard,
	}, nil
}

func (r *SessionRepo) GetCookies(ctx context.Context, puuid string) ([]*http.Cookie, error) {
	val, err := r.client.Get(ctx, cookiesKey(puuid)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCookiesNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get cookies: %w", err)
	}

	var data cookieData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal cookies: %w", err)
	}
	return data.Cookies, nil
}

func (r *SessionRepo) DeleteSession(ctx context.Context, puuid string) error {
	return r.client.Del(ctx, sessionKey(puuid), cookiesKey(puuid)).Err()
}

func sessionKey(puuid string) string { return "session:" + puuid }
func cookiesKey(puuid string) string { return "cookies:" + puuid }
