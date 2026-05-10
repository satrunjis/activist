package postgres

import (
	"context"
	"fmt"

	appeventlog "activist-base/src/application/eventlog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) Begin(ctx context.Context) (appeventlog.Transaction, error) {
	if m == nil || m.pool == nil {
		return nil, fmt.Errorf("tx manager requires pgx pool")
	}
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return &pgxTransaction{tx: tx}, nil
}

type pgxTransaction struct {
	tx pgx.Tx
}

func (t *pgxTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func unwrapPGXTx(tx appeventlog.Transaction) (pgx.Tx, error) {
	typed, ok := tx.(*pgxTransaction)
	if !ok || typed == nil || typed.tx == nil {
		return nil, fmt.Errorf("unsupported transaction type %T", tx)
	}
	return typed.tx, nil
}
