package repository

import (
	"context"

	"github.com/sawthorn/merch-wallet/internal/database/transactions"
)

type InventoryRepository struct {
	executor *transactions.DBExecutor
}

func NewInventoryRepository(ex *transactions.DBExecutor) *InventoryRepository {
	return &InventoryRepository{executor: ex}
}

func (pr *InventoryRepository) AddItemToUserInventory(ctx context.Context, userID int, itemID int) error {
	ex := pr.executor.For(ctx)
	_, err := ex.ExecContext(ctx, `
		INSERT INTO inventory(user_id, item_id, quantity)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, item_id)
		DO UPDATE SET
			quantity = inventory.quantity + 1
		`, userID, itemID)
	return err
}
