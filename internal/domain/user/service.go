package user

import (
	"ValorantAPI/internal/pkg/hash"
	"context"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, user *User) error {
	passwordHash, err := hash.GeneratePasswordHash(user.Password)
	if err != nil {
		return err
	}
	user.Password = passwordHash
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) GetUserByLogin(ctx context.Context, user *User) error {
	return s.repo.GetUserByLogin(ctx, user)
}

func (s *Service) GetUserByID(ctx context.Context, user *User) error {
	return s.repo.GetUserByID(ctx, user)
}
