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
	trip, err := scanTrip(
		r.pool.QueryRow(
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
		),
	)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("get trip: %w", err)
	}

	return trip, nil
}

func (r *TripRepoTx) GetForUpdateByID(ctx context.Context, tripId uuid.UUID) (domain.Trip, error) {
	trip, err := scanTrip(
		r.tx.QueryRow(
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
			WHERE id = $1
			FOR UPDATE`,
			tripId,
		),
	)
	if err != nil {
		return domain.Trip{}, fmt.Errorf("get trip for update: %w", err)
	}

	return trip, nil
}

func (r *TripRepoTx) UpdateStatus(ctx context.Context, tripId uuid.UUID, status domain.TripStatus) error {
	commandTag, err := r.tx.Exec(
		ctx,
		`UPDATE trips
			SET status = $2
			WHERE id = $1`,
		tripId,
		string(status),
	)
	if err != nil {
		return fmt.Errorf("update trip status: %w", mapPostgresError(err))
	}
	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *TripRepoTx) AppendHistory(ctx context.Context, history domain.TripHistory) error {
	historyEntity := toTripHistoryEntity(history)
	if err := insertTripHistory(ctx, r.tx, historyEntity); err != nil {
		return fmt.Errorf("append trip history: %w", err)
	}

	return nil
}

func scanTrip(row pgx.Row) (domain.Trip, error) {
	var trip tripEntity
	err := row.Scan(
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
		return domain.Trip{}, mapPostgresError(err)
	}

	return toDomainTrip(trip), nil
}
