package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TripRepoTx struct {
	tx pgx.Tx
}

func newTripRepoTx(tx pgx.Tx) *TripRepoTx {
	return &TripRepoTx{tx: tx}
}

func (r *RepoPg) WithinTripTx(ctx context.Context, fn func(ctx context.Context, trips *TripRepoTx) error) error {
	return tx(ctx, r.pool, func(txBegin pgx.Tx) error {
		return fn(ctx, newTripRepoTx(txBegin))
	})
}
