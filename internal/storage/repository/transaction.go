package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tx(ctx context.Context, pool *pgxpool.Pool, block func(tx pgx.Tx) error) error {
	txBegin, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", mapPostgresError(err))
	}
	defer func() {
		_ = txBegin.Rollback(ctx)
	}()

	if err := block(txBegin); err != nil {
		return fmt.Errorf("transaction block: %w", err)
	}

	if err := txBegin.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", mapPostgresError(err))
	}

	return nil
}
