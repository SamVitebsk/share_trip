package repository

import (
	"context"
	"errors"
	"fmt"

	"share_trip/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func (r *RepoPg) Create(ctx context.Context, trip domain.Trip, history domain.TripHistory) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin create trip: %w", mapPostgresError(err))
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := insertTrip(ctx, tx, trip); err != nil {
		return err
	}
	if err := insertTripHistory(ctx, tx, history); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create trip: %w", mapPostgresError(err))
	}

	return nil
}

func insertTrip(ctx context.Context, exec executor, trip domain.Trip) error {
	_, err := exec.Exec(
		ctx,
		`INSERT INTO trips(
                  id,
                  driver_id,
                  from_point,
                  to_point,
                  departure_time,
                  seats,
                  status,
                  created_at
			  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		trip.ID,
		trip.DriverID,
		trip.FromPoint,
		trip.ToPoint,
		trip.DepartureTime,
		trip.Seats,
		trip.Status,
		trip.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create trip: %w", mapPostgresError(err))
	}

	return nil
}

func insertTripHistory(ctx context.Context, exec executor, history domain.TripHistory) error {
	_, err := exec.Exec(
		ctx,
		`INSERT INTO trip_history(
                  id,
                  trip_id,
                  from_status,
                  to_status,
                  created_at
			  ) VALUES ($1, $2, $3, $4, $5)`,
		history.ID,
		history.TripID,
		nullableTripStatus(history.FromStatus),
		history.ToStatus,
		history.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create trip history: %w", mapPostgresError(err))
	}

	return nil
}

func nullableTripStatus(status *domain.TripStatus) any {
	if status == nil {
		return nil
	}

	return *status
}

func (r *RepoPg) GetByID(ctx context.Context, tripId uuid.UUID) (domain.Trip, error) {
	var trip domain.Trip
	err := r.pool.QueryRow(
		ctx,
		`SELECT 
    			id,
    			driver_id,
    			from_point,
    			to_point,
    			departure_time,
    			seats,
    			status,
    			created_at
			FROM trips
			WHERE id = $1`,
		tripId,
	).Scan(
		&trip.ID,
		&trip.DriverID,
		&trip.FromPoint,
		&trip.ToPoint,
		&trip.DepartureTime,
		&trip.Seats,
		&trip.Status,
		&trip.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Trip{}, ErrNotFound
	}
	if err != nil {
		return domain.Trip{}, fmt.Errorf("get trip: %w", mapPostgresError(err))
	}

	return trip, nil
}
