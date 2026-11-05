package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func (tr *TransferRepository) GetReceivedHistory(ctx context.Context, userID int) ([]TransferHistoryEntry, error) {
	ex := tr.executor.For(ctx)
	rows, err := ex.QueryContext(ctx, `
		SELECT username, amount, created_at
		FROM transfers t JOIN users u ON t.from_user_id = u.id 
		WHERE to_user_id = $1
		`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return []TransferHistoryEntry{}, nil
	}
	if err != nil {
		return nil, err
	}

	recievedTr := []TransferHistoryEntry{}
	for rows.Next() {
		transfer := TransferHistoryEntry{}
		var tsStr string
		err := rows.Scan(&transfer.username, &transfer.amount, &tsStr)
		if err != nil {
			return nil, err
		}
		transfer.createdAt, err = time.Parse(time.RFC3339Nano, tsStr)
		if err != nil {
			return nil, err
		}
		recievedTr = append(recievedTr, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return recievedTr, nil
}

func (tr *TransferRepository) GetSentHistory(ctx context.Context, userID int) ([]TransferHistoryEntry, error) {
	ex := tr.executor.For(ctx)
	rows, err := ex.QueryContext(ctx, `
		SELECT username, amount, created_at
		FROM transfers t JOIN users u ON t.to_user_id = u.id 
		WHERE from_user_id = $1
		`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return []TransferHistoryEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	sentTr := []TransferHistoryEntry{}
	for rows.Next() {
		transfer := TransferHistoryEntry{}
		var tsStr string
		err := rows.Scan(&transfer.username, &transfer.amount, &tsStr)
		if err != nil {
			return nil, err
		}
		transfer.createdAt, err = time.Parse(time.RFC3339Nano, tsStr)
		if err != nil {
			return nil, err
		}
		sentTr = append(sentTr, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sentTr, nil
}
