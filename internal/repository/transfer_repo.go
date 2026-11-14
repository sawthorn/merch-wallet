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

func (tr *TransferRepository) CreateTransfer(ctx context.Context, from int, to int, amount int) error {
	ex := tr.executor.For(ctx)
	_, err := ex.ExecContext(ctx, `
		INSERT INTO transfers(from_user_id, to_user_id, amount)
		VALUES($1, $2, $3)
		`, from, to, amount)
	return mapPGErrorToRepoError(err)
}

func (tr *TransferRepository) GetReceivedHistory(ctx context.Context, userID int) ([]TransferHistoryEntry, error) {
	query := `SELECT username, amount, created_at
		FROM transfers t JOIN users u ON t.from_user_id = u.id 
		WHERE to_user_id = $1
		ORDER BY created_at DESC
		`
	return tr.getHistory(ctx, userID, query)
}

func (tr *TransferRepository) GetSentHistory(ctx context.Context, userID int) ([]TransferHistoryEntry, error) {
	query := `SELECT username, amount, created_at
		FROM transfers t JOIN users u ON t.to_user_id = u.id 
		WHERE from_user_id = $1
		ORDER BY created_at DESC
		`
	return tr.getHistory(ctx, userID, query)
}

func (tr *TransferRepository) getHistory(ctx context.Context, userID int, query string) ([]TransferHistoryEntry, error) {
	ex := tr.executor.For(ctx)
	rows, err := ex.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trHistory := []TransferHistoryEntry{}
	for rows.Next() {
		transfer := TransferHistoryEntry{}
		if err := rows.Scan(
			&transfer.username,
			&transfer.amount,
			&transfer.createdAt,
		); err != nil {
			return trHistory, err
		}
		trHistory = append(trHistory, transfer)
	}
	if err := rows.Err(); err != nil {
		return trHistory, err
	}
	return trHistory, nil
}
