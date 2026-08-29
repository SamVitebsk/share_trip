package repository

import (
	"context"
	"fmt"

	"share_trip/internal/observability/metrics"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepoPg struct {
	pool    *pgxpool.Pool
	metrics *metrics.Metrics
}

func NewRepoPg(pool *pgxpool.Pool, metrics *metrics.Metrics) *RepoPg {
	return &RepoPg{
		pool:    pool,
		metrics: metrics,
	}
}

func (r *RepoPg) Ping(ctx context.Context) error {
	err := r.pool.Ping(ctx)

	if err != nil {
		return fmt.Errorf("r.pool.Ping: %w", err)
	}

	return nil
}
