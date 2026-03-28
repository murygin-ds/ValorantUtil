package postgres

import "errors"

var (
	ErrLoginAlreadyTaken = errors.New("login is already taken")
	ErrUserNotFound      = errors.New("user not found")
)
