package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"share_trip/internal/observability/logctx"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tx(ctx context.Context, pool *pgxpool.Pool, block func(ctx context.Context, tx pgx.Tx) error) (err error) {
	logger := logctx.Logger(ctx).With(
		slog.String("layer", "transaction"),
	)

	logger.InfoContext(ctx, "открытие транзакции начато")

	txBegin, err := pool.Begin(ctx)
	if err != nil {
		logger.ErrorContext(
			ctx,
			"открытие транзакции не выполнено",
			slog.Any("error", err),
		)
		return fmt.Errorf("begin transaction: %w", mapPostgresError(err))
	}
	defer func() {
		if err != nil {
			if rollbackErr := txBegin.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				logger.ErrorContext(
					ctx,
					"rollback транзакции не выполнен",
					slog.Any("error", rollbackErr),
				)
			}

			logger.ErrorContext(
				ctx,
				"транзакция завершилась с ошибкой",
				slog.Any("error", err),
			)
			return
		}
	}()

	if err := block(ctx, txBegin); err != nil {
		logger.ErrorContext(
			ctx,
			"блок транзакции завершился с ошибкой",
			slog.Any("error", err),
		)
		return fmt.Errorf("transaction block: %w", err)
	}

	if err := txBegin.Commit(ctx); err != nil {
		logger.ErrorContext(
			ctx,
			"commit транзакции не выполнен",
			slog.Any("error", err),
		)
		return fmt.Errorf("commit transaction: %w", mapPostgresError(err))
	}

	logger.InfoContext(ctx, "commit транзакции выполнен")

	return nil
}
