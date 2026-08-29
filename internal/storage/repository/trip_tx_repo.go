package repository

import (
	"context"

	"share_trip/internal/observability/metrics"

	"github.com/jackc/pgx/v5"
)

type TripRepoTx struct {
	tx      pgx.Tx
	metrics *metrics.Metrics
}

func newTripRepoTx(tx pgx.Tx, metrics *metrics.Metrics) *TripRepoTx {
	return &TripRepoTx{
		tx:      tx,
		metrics: metrics,
	}
}

func (r *RepoPg) WithinTripTx(ctx context.Context, fn func(ctx context.Context, trips *TripRepoTx) error) error {
	return tx(ctx, r.pool, func(ctx context.Context, txBegin pgx.Tx) error {
		return fn(ctx, newTripRepoTx(txBegin, r.metrics))
	})
}
