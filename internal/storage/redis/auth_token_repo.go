package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const refreshTokenTTL = 7 * 24 * time.Hour

type AuthTokenRepo struct {
	client *redis.Client
}

func NewAuthTokenRepo(client *redis.Client) *AuthTokenRepo {
	return &AuthTokenRepo{client: client}
}

func GenerateRefreshTokenUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (r *AuthTokenRepo) SaveRefreshToken(ctx context.Context, uuid string, userID int64) error {
	return r.client.Set(ctx, refreshTokenKey(uuid), strconv.FormatInt(userID, 10), refreshTokenTTL).Err()
}

func (r *AuthTokenRepo) GetRefreshToken(ctx context.Context, uuid string) (int64, error) {
	val, err := r.client.Get(ctx, refreshTokenKey(uuid)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrRefreshTokenNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get refresh token: %w", err)
	}
	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse user id: %w", err)
	}
	return userID, nil
}

func (r *AuthTokenRepo) DeleteRefreshToken(ctx context.Context, uuid string) error {
	return r.client.Del(ctx, refreshTokenKey(uuid)).Err()
}

func refreshTokenKey(uuid string) string { return "refresh:" + uuid }
