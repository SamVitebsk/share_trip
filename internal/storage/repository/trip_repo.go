package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"share_trip/internal/domain"
	"share_trip/internal/observability/logctx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func (r *RepoPg) Create(ctx context.Context, trip domain.Trip, history domain.TripHistory) error {
	tracer := otel.Tracer("TripRepository")
	ctx, span := tracer.Start(ctx, "TripRepository.Create")
	defer span.End()
	span.SetAttributes(
		attribute.String("operation", "trip_create"),
		attribute.String("trip_id", trip.ID.String()),
		attribute.String("driver_id", trip.DriverID.String()),
		attribute.String("status", string(trip.Status)),
	)

	started := time.Now()
	result := repositoryMetricResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues("trip_create", result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues("trip_create", result).
			Observe(time.Since(started).Seconds())
	}()

	err := tx(ctx, r.pool, func(ctx context.Context, txBegin pgx.Tx) error {
		tripEntity := toTripEntity(trip)
		if err := insertTrip(ctx, txBegin, tripEntity); err != nil {
			return err
		}

		historyEntity := toTripHistoryEntity(history)
		if err := insertTripHistory(ctx, txBegin, historyEntity); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		result = repositoryMetricResultFromError(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func insertTrip(ctx context.Context, exec executor, tripEntity tripEntity) error {
	tracer := otel.Tracer("TripRepository")
	ctx, span := tracer.Start(ctx, "TripRepository.InsertTrip")
	defer span.End()
	span.SetAttributes(
		attribute.String("operation", "trip_insert"),
		attribute.String("trip_id", tripEntity.ID.String()),
		attribute.String("driver_id", tripEntity.DriverID.String()),
		attribute.String("status", tripEntity.Status),
	)

	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "Create"),
	)

	logger.InfoContext(ctx, "insert поездки начат")

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.ErrorContext(
			ctx,
			"insert поездки не выполнен",
			slog.Any("error", err),
		)
		return fmt.Errorf("create trip: %w", mapPostgresError(err))
	}

	logger.InfoContext(ctx, "insert поездки выполнен")

	return nil
}

func insertTripHistory(ctx context.Context, exec executor, historyEntity tripHistoryEntity) error {
	tracer := otel.Tracer("TripRepository")
	ctx, span := tracer.Start(ctx, "TripRepository.InsertTripHistory")
	defer span.End()

	spanAttributes := []attribute.KeyValue{
		attribute.String("operation", "trip_history_insert"),
		attribute.String("history_id", historyEntity.ID.String()),
		attribute.String("trip_id", historyEntity.TripID.String()),
		attribute.String("to_status", historyEntity.ToStatus),
	}
	if historyEntity.FromStatus != nil {
		spanAttributes = append(spanAttributes, attribute.String("from_status", *historyEntity.FromStatus))
	}
	span.SetAttributes(spanAttributes...)

	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "CreateHistory"),
		slog.String("history_id", historyEntity.ID.String()),
	)

	logger.InfoContext(ctx, "insert истории поездки начат")

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.ErrorContext(
			ctx,
			"insert истории поездки не выполнен",
			slog.Any("error", err),
		)
		return fmt.Errorf("create trip history: %w", mapPostgresError(err))
	}

	logger.InfoContext(ctx, "insert истории поездки выполнен")

	return nil
}

func (r *RepoPg) GetByID(ctx context.Context, tripId uuid.UUID) (domain.Trip, error) {
	tracer := otel.Tracer("TripRepository")
	ctx, span := tracer.Start(ctx, "TripRepository.GetByID")
	defer span.End()
	span.SetAttributes(
		attribute.String("operation", "trip_get_by_id"),
		attribute.String("trip_id", tripId.String()),
	)

	started := time.Now()
	result := repositoryMetricResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues("trip_get_by_id", result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues("trip_get_by_id", result).
			Observe(time.Since(started).Seconds())
	}()

	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "GetByID"),
	)

	logger.InfoContext(ctx, "select поездки по id начат")

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
		result = repositoryMetricResultFromError(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, ErrNotFound) {
			logger.WarnContext(
				ctx,
				"select поездки по id не выполнен: поездка не найдена",
				slog.Any("error", err),
			)
			return domain.Trip{}, fmt.Errorf("get trip: %w", err)
		}

		logger.ErrorContext(
			ctx,
			"select поездки по id не выполнен",
			slog.Any("error", err),
		)
		return domain.Trip{}, fmt.Errorf("get trip: %w", err)
	}

	logger.InfoContext(
		ctx,
		"select поездки по id выполнен",
		slog.String("driver_id", trip.DriverID.String()),
		slog.String("status", string(trip.Status)),
	)
	span.SetAttributes(
		attribute.String("driver_id", trip.DriverID.String()),
		attribute.String("trip_id", tripId.String()),
		attribute.String("status", string(trip.Status)),
	)

	return trip, nil
}

func (r *TripRepoTx) GetForUpdateByID(ctx context.Context, tripId uuid.UUID) (domain.Trip, error) {
	tracer := otel.Tracer("TripRepository")
	ctx, span := tracer.Start(ctx, "TripRepository.GetForUpdateByID")
	defer span.End()
	span.SetAttributes(
		attribute.String("operation", "trip_get_for_update"),
		attribute.String("trip_id", tripId.String()),
	)

	started := time.Now()
	result := repositoryMetricResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues("trip_get_for_update", result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues("trip_get_for_update", result).
			Observe(time.Since(started).Seconds())
	}()

	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "GetForUpdateByID"),
	)

	logger.InfoContext(ctx, "select поездки для обновления начат")

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
		result = repositoryMetricResultFromError(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.ErrorContext(
			ctx,
			"select поездки для обновления не выполнен",
			slog.Any("error", err),
		)
		return domain.Trip{}, fmt.Errorf("get trip for update: %w", err)
	}

	logger.InfoContext(
		ctx,
		"select поездки для обновления выполнен",
		slog.String("status", string(trip.Status)),
	)
	span.SetAttributes(
		attribute.String("driver_id", trip.DriverID.String()),
		attribute.String("status", string(trip.Status)),
	)

	return trip, nil
}

func (r *TripRepoTx) UpdateStatus(ctx context.Context, tripId uuid.UUID, status domain.TripStatus) (domain.Trip, error) {
	tracer := otel.Tracer("TripRepository")
	ctx, span := tracer.Start(ctx, "TripRepository.UpdateStatus")
	defer span.End()
	span.SetAttributes(
		attribute.String("operation", "trip_update_status"),
		attribute.String("trip_id", tripId.String()),
		attribute.String("to_status", string(status)),
	)

	started := time.Now()
	result := repositoryMetricResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues("trip_update_status", result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues("trip_update_status", result).
			Observe(time.Since(started).Seconds())
	}()

	logger := logctx.Logger(ctx).With(
		slog.String("layer", "repository"),
		slog.String("repository", "TripRepository"),
		slog.String("operation", "UpdateStatus"),
		slog.String("to_status", string(status)),
	)

	logger.InfoContext(ctx, "update статуса поездки начат")

	trip, err := scanTrip(r.tx.QueryRow(
		ctx,
		`UPDATE trips
			SET status = $2
			WHERE id = $1
			RETURNING
				id,
				driver_id,
				from_point,
				to_point,
				departure_time,
				seats,
				status,
				created_at`,
		tripId,
		string(status),
	))
	if err != nil {
		result = repositoryMetricResultFromError(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.ErrorContext(
			ctx,
			"update статуса поездки не выполнен",
			slog.Any("error", err),
		)
		return domain.Trip{}, fmt.Errorf("update trip status: %w", mapPostgresError(err))
	}

	logger.InfoContext(ctx, "update статуса поездки выполнен")
	span.SetAttributes(
		attribute.String("driver_id", trip.DriverID.String()),
		attribute.String("status", string(trip.Status)),
	)

	return trip, nil
}

func (r *TripRepoTx) CreateHistory(ctx context.Context, history domain.TripHistory) error {
	tracer := otel.Tracer("TripRepository")
	ctx, span := tracer.Start(ctx, "TripRepository.CreateHistory")
	defer span.End()

	spanAttributes := []attribute.KeyValue{
		attribute.String("operation", "trip_create_history"),
		attribute.String("history_id", history.ID.String()),
		attribute.String("trip_id", history.TripID.String()),
		attribute.String("to_status", string(history.ToStatus)),
	}
	if history.FromStatus != nil {
		spanAttributes = append(spanAttributes, attribute.String("from_status", string(*history.FromStatus)))
	}
	span.SetAttributes(spanAttributes...)

	started := time.Now()
	result := repositoryMetricResultSuccess

	defer func() {
		r.metrics.RepositoryQueryTotal.WithLabelValues("trip_create_history", result).Inc()
		r.metrics.RepositoryQueryDuration.WithLabelValues("trip_create_history", result).
			Observe(time.Since(started).Seconds())
	}()

	historyEntity := toTripHistoryEntity(history)
	if err := insertTripHistory(ctx, r.tx, historyEntity); err != nil {
		result = repositoryMetricResultFromError(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
