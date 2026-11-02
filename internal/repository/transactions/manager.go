// Package transactions provides utility abstractions for transaction management
package transactions

import (
	"context"
	"database/sql"
	"errors"
)

type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (t *TransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	nested := false
	tx := &sql.Tx{}

	if transaction, ok := transactionFromContext(ctx); ok {
		nested = true
		tx = transaction
	} else {
		tx, err = t.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return err
		}
		defer func() {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				err = rollbackErr
			}
		}()
	}

	trCtx := context.WithValue(ctx, txContextKey{}, tx)
	if err := fn(trCtx); err != nil {
		return err
	}

	if nested {
		return nil
	}
	return tx.Commit()
}
