package valorant

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateAccount(ctx context.Context, account *Account) error {
	return s.repo.CreateAccount(ctx, account)
}

func (s *Service) GetAccountsList(ctx context.Context, userID, limit, offset int) ([]Account, error) {
	return s.repo.GetAccountsList(ctx, userID, limit, offset)
}
