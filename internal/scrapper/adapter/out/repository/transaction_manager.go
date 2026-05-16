package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type txKey string

var ctxWithTx = txKey("tx")

type TransactionManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

func (m *TransactionManager) WithTx(ctx context.Context, callback func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot start transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()
	txctx := putTxToContext(ctx, tx)
	err = callback(txctx)
	if err != nil {
		return fmt.Errorf("transaction callback error: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func ExtractTxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx := ctx.Value(ctxWithTx)

	if t, ok := tx.(*sql.Tx); ok {
		return t, true
	}

	return nil, false
}

func putTxToContext(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, ctxWithTx, tx)
}

func GetExecutor(ctx context.Context, db *sql.DB) Executor {
	tx, ok := ExtractTxFromContext(ctx)
	if ok {
		return tx
	}
	return db
}
