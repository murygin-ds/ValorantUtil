package valorant

import (
	"context"
)

type Repository interface {
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountsList(ctx context.Context, userID, limit, offset int) ([]Account, error)
}
