package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tx(ctx context.Context, pool *pgxpool.Pool, block func(tx pgx.Tx) error) (err error) {
	txBegin, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", mapPostgresError(err))
	}
	defer func() {
		if err != nil {
			if rollbackErr := txBegin.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				slog.ErrorContext(ctx, "rollback transaction failed", "error", rollbackErr)
			}

			slog.ErrorContext(ctx, "transaction failed", "error", err)
			return
		}

		slog.InfoContext(ctx, "transaction committed")
	}()

	if err := block(txBegin); err != nil {
		return fmt.Errorf("transaction block: %w", err)
	}

	if err := txBegin.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", mapPostgresError(err))
	}

	return nil
}
