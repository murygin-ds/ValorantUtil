package user

import (
	"context"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByLogin(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, user *User) error
}
