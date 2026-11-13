package repository

import (
	"context"

	"github.com/sawthorn/merch-wallet/internal/database/transactions"
)

type PurchaseRepository struct {
	executor *transactions.DBExecutor
}

func NewPurchaseRepository(ex *transactions.DBExecutor) *PurchaseRepository {
	return &PurchaseRepository{executor: ex}
}

func (pr *PurchaseRepository) CreatePurchase(ctx context.Context, userID int, itemID int) error {
	ex := pr.executor.For(ctx)
	_, err := ex.ExecContext(ctx, `
		INSERT INTO purchases(user_id, item_id)
		VALUES($1, $2)
		`, userID, itemID)
	return mapPGErrorToRepoError(err)
}
