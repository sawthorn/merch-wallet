package repository

import (
	"context"
	"errors"

	"github.com/lib/pq"
	"github.com/sawthorn/merch-wallet/internal/database/transactions"
)

type InventoryRepository struct {
	executor *transactions.DBExecutor
}

func NewInventoryRepository(ex *transactions.DBExecutor) *InventoryRepository {
	return &InventoryRepository{executor: ex}
}

func mapPGErrorToRepoError(err error) error {
	if pgErr, ok := errors.AsType[*pq.Error](err); ok {
		if pgErr.Code == "23503" { // FK constraint violation
			if pgErr.Constraint == "inventory_user_id_fkey" {
				return ErrUserNotFound
			}
			if pgErr.Constraint == "inventory_item_id_fkey" {
				return ErrItemNotFound
			}
		}
	}
	return err
}

func (ir *InventoryRepository) AddItemToUserInventory(ctx context.Context, userID int, itemID int) error {
	ex := ir.executor.For(ctx)
	_, err := ex.ExecContext(ctx, `
		INSERT INTO inventory(user_id, item_id, quantity)
		VALUES ($1, $2, 1)
		ON CONFLICT (user_id, item_id)
		DO UPDATE SET
			quantity = inventory.quantity + 1
		`, userID, itemID)
	return mapPGErrorToRepoError(err)
}

func (ir *InventoryRepository) GetUserInventory(ctx context.Context, userID int) ([]InventoryItem, error) {
	ex := ir.executor.For(ctx)
	rows, err := ex.QueryContext(ctx, `
		SELECT name, quantity
		FROM inventory inv JOIN items ON inv.item_id = items.id
		WHERE user_id = $1
		ORDER BY name
		`, userID)
	if err != nil {
		return nil, mapPGErrorToRepoError(err)
	}

	inv := []InventoryItem{}
	for rows.Next() {
		item := InventoryItem{}
		rows.Scan(&item.name, &item.quantity)
		inv = append(inv, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapPGErrorToRepoError(err)
	}

	return inv, nil
}
