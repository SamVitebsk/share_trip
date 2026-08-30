package repository

import (
	"context"

	"share_trip/internal/observability/metrics"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	tracer := otel.Tracer("TripRepository")
	ctx, span := tracer.Start(ctx, "TripRepository.WithinTripTx")
	defer span.End()
	span.SetAttributes(attribute.String("operation", "trip_transaction"))

	err := tx(ctx, r.pool, func(ctx context.Context, txBegin pgx.Tx) error {
		return fn(ctx, newTripRepoTx(txBegin, r.metrics))
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
