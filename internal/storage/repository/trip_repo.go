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

	tripEntity := toTripEntity(trip)
	if err := insertTrip(ctx, tx, tripEntity); err != nil {
		return err
	}

	historyEntity := toTripHistoryEntity(history)
	if err := insertTripHistory(ctx, tx, historyEntity); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit create trip: %w", mapPostgresError(err))
	}

	return nil
}

func insertTrip(ctx context.Context, exec executor, tripEntity tripEntity) error {
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
		tripEntity.ID,
		tripEntity.DriverID,
		tripEntity.FromPoint,
		tripEntity.ToPoint,
		tripEntity.DepartureTime,
		tripEntity.Seats,
		tripEntity.Status,
		tripEntity.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create trip: %w", mapPostgresError(err))
	}

	return nil
}

func insertTripHistory(ctx context.Context, exec executor, historyEntity tripHistoryEntity) error {
	_, err := exec.Exec(
		ctx,
		`INSERT INTO trip_history(
                  id,
                  trip_id,
                  from_status,
                  to_status,
                  created_at
			  ) VALUES ($1, $2, $3, $4, $5)`,
		historyEntity.ID,
		historyEntity.TripID,
		historyEntity.FromStatus,
		historyEntity.ToStatus,
		historyEntity.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create trip history: %w", mapPostgresError(err))
	}

	return nil
}

func (r *RepoPg) GetByID(ctx context.Context, tripId uuid.UUID) (domain.Trip, error) {
	var trip tripEntity
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

	return toDomainTrip(trip), nil
}
