package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sawthorn/merch-wallet/internal/database/transactions"
)

var ErrItemNotFound = errors.New("item not found")

type ItemRepository struct {
	executor *transactions.DBExecutor
}

func NewItemRepository(executor *transactions.DBExecutor) *ItemRepository {
	return &ItemRepository{
		executor: executor,
	}
}

func (ir *ItemRepository) GetItemByName(ctx context.Context, itemName string) (*Item, error) {
	ex := ir.executor.For(ctx)
	item := &Item{}
	err := ex.QueryRowContext(ctx, `
		SELECT id, name, cost
		FROM items
		WHERE name = $1
		`, itemName).Scan(
		&item.id,
		&item.name,
		&item.cost,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}
