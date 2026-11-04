package transactions

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(ctx context.Context, sql string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, sql string, args ...any) *sql.Row
}

type DBExecutor struct {
	defaultExecutor *sql.DB
}

func NewDBExecutor(db *sql.DB) *DBExecutor {
	return &DBExecutor{defaultExecutor: db}
}

func (e *DBExecutor) For(ctx context.Context) DBTX {
	if tx, ok := transactionFromContext(ctx); ok {
		return tx
	}
	return e.defaultExecutor
}

type txContextKey struct{}

func transactionFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*sql.Tx)
	return tx, ok && tx != nil
}
