package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepoPg struct {
	pool *pgxpool.Pool
}

func NewRepoPg(pool *pgxpool.Pool) *RepoPg {
	return &RepoPg{pool: pool}
}

func (r *RepoPg) Ping(ctx context.Context) error {
	err := r.pool.Ping(ctx)

	if err != nil {
		return fmt.Errorf("r.pool.Ping: %w", err)
	}

	return nil
}
