package repository

import (
	"context"

	"github.com/sawthorn/merch-wallet/internal/database/transactions"
)

type TransferRepository struct {
	executor *transactions.DBExecutor
}

func NewTransferRepository(executor *transactions.DBExecutor) *TransferRepository {
	return &TransferRepository{
		executor: executor,
	}
}

func (tr *TransferRepository) CreateTransfer(ctx context.Context, transfer *Transfer) error {
	ex := tr.executor.For(ctx)
	_, err := ex.ExecContext(ctx, `
		INSERT INTO transfers(from_user_id, to_user_id, amount)
		VALUES($1, $2, $3)
		`, transfer.fromID, transfer.toID, transfer.amount)
	return err
}
