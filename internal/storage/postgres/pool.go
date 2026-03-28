package postgres

import (
	"ValorantAPI/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	pgxCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pgxCfg.MinConns = int32(cfg.Pool.MinConnections)
	pgxCfg.MaxConns = int32(cfg.Pool.MaxConnections)
	pgxCfg.MaxConnLifetime = time.Duration(cfg.Pool.MaxConnectionLifetime)
	pgxCfg.MaxConnIdleTime = time.Duration(cfg.Pool.MaxConnectionIdleTime)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return pool, nil
}
