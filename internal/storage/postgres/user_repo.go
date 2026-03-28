package postgres

import (
	"ValorantAPI/internal/domain/user"
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		pool: pool,
	}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *user.User) error {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at`
	if err := r.pool.QueryRow(ctx, query, user.Login, user.Password).Scan(&user.ID, &user.CreatedAt); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return ErrLoginAlreadyTaken
		}
		return err
	}
	return nil
}

func (r *UserRepo) GetUserByLogin(ctx context.Context, user *user.User) error {
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`
	if err := r.pool.QueryRow(ctx, query, user.Login).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, user *user.User) error {
	query := `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`
	if err := r.pool.QueryRow(ctx, query, user.ID).Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}
