// Package transactions provides utility abstractions for transaction management
package transactions

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"github.com/lib/pq"
)

const (
	maxTxAttempts = 3
	baseBackoff   = 25 * time.Millisecond
	maxBackoff    = 500 * time.Millisecond
)

var ErrRetryInsideNestedTx = errors.New("cannot use retry inside nested transaction")

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

func (t *TransactionManager) RunInTxWithRetry(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := transactionFromContext(ctx); ok {
		return ErrRetryInsideNestedTx
	}

	for attempt := 1; attempt <= maxTxAttempts; attempt++ {
		if err = ctx.Err(); err != nil {
			return err
		}

		err = t.RunInTx(ctx, fn)
		if err == nil {
			return nil
		}

		if !isDeadlock(err) {
			return err
		}
		// no need to fire timer for nothing - there wont be next iteration
		if attempt == maxTxAttempts {
			return err
		}

		backoff := baseBackoff << (attempt - 1)
		backoff = min(backoff, maxBackoff)
		jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
		timer := time.NewTimer(backoff + jitter)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}

func isDeadlock(err error) bool {
	var pgErr *pq.Error

	return errors.As(err, &pgErr) &&
		pgErr.Code == "40P01"
}
