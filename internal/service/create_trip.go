package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"share_trip/internal/domain"
	"share_trip/internal/observability/logctx"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type CreateTripCommand struct {
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
}

type CreateTripResult struct {
	TripID uuid.UUID
}

func (s *TripService) CreateTrip(ctx context.Context, cmd CreateTripCommand) (CreateTripResult, error) {
	tracer := otel.Tracer("TripService")
	ctx, span := tracer.Start(ctx, "TripService.CreateTrip")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "create_trip"),
		attribute.String("driver_id", cmd.DriverID.String()),
	)

	started := time.Now()
	result := metricResultSuccess

	defer func() {
		s.metrics.TripCreateTotal.WithLabelValues(result).Inc()
		s.metrics.TripCreateDuration.WithLabelValues(result).
			Observe(time.Since(started).Seconds())
	}()

	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "CreateTrip"),
	)

	logger.Info("создание поездки в service начато")

	now := time.Now()
	cmd = normalizeCreateTripCommand(cmd)
	if err := validateCreateTripCommand(cmd, now); err != nil {
		result = metricResultFromError(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn(
			"создание поездки не выполнено: ошибка валидации",
			slog.Any("error", err),
		)
		return CreateTripResult{}, err
	}

	trip := domain.Trip{
		ID:            uuid.New(),
		DriverID:      cmd.DriverID,
		FromPoint:     cmd.FromPoint,
		ToPoint:       cmd.ToPoint,
		DepartureTime: cmd.DepartureTime,
		Seats:         cmd.Seats,
		Status:        domain.TripStatusDraft,
		CreatedAt:     now,
	}
	history := domain.TripHistory{
		ID:         uuid.New(),
		TripID:     trip.ID,
		FromStatus: nil,
		ToStatus:   trip.Status,
		CreatedAt:  now,
	}
	span.SetAttributes(
		attribute.String("trip_id", trip.ID.String()),
		attribute.String("driver_id", trip.DriverID.String()),
		attribute.String("status", string(trip.Status)),
	)

	ctx = logctx.WithLogger(ctx, logctx.Logger(ctx).With(
		slog.String("trip_id", trip.ID.String()),
		slog.String("driver_id", trip.DriverID.String()),
	))

	err := s.tripRepository.Create(ctx, trip, history)
	if err != nil {
		result = metricResultFromError(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error(
			"создание поездки не выполнено: ошибка repository",
			slog.Any("error", err),
		)
		return CreateTripResult{}, err
	}

	logger.Info(
		"создание поездки в service завершено",
		slog.String("trip_id", trip.ID.String()),
	)

	return CreateTripResult{TripID: trip.ID}, nil
}

func normalizeCreateTripCommand(cmd CreateTripCommand) CreateTripCommand {
	cmd.FromPoint = strings.TrimSpace(cmd.FromPoint)
	cmd.ToPoint = strings.TrimSpace(cmd.ToPoint)
	return cmd
}

func validateCreateTripCommand(cmd CreateTripCommand, now time.Time) error {
	var validationErrors []FieldError

	if cmd.DriverID == uuid.Nil {
		validationErrors = append(validationErrors, FieldError{
			Field:   "driverId",
			Message: "ID водителя обязателен",
		})
	}
	if cmd.FromPoint == "" {
		validationErrors = append(validationErrors, FieldError{
			Field:   "fromPoint",
			Message: "пункт отправления обязателен",
		})
	}
	if cmd.ToPoint == "" {
		validationErrors = append(validationErrors, FieldError{
			Field:   "toPoint",
			Message: "пункт назначения обязателен",
		})
	}
	if cmd.FromPoint != "" && cmd.ToPoint != "" && cmd.FromPoint == cmd.ToPoint {
		validationErrors = append(validationErrors, FieldError{
			Field:   "route",
			Message: "пункт отправления и пункт назначения должны отличаться",
		})
	}
	if cmd.DepartureTime.IsZero() {
		validationErrors = append(validationErrors, FieldError{
			Field:   "departureTime",
			Message: "время отправления обязательно",
		})
	} else if !cmd.DepartureTime.After(now) {
		validationErrors = append(validationErrors, FieldError{
			Field:   "departureTime",
			Message: "время отправления должно быть в будущем",
		})
	}
	if cmd.Seats <= 0 {
		validationErrors = append(validationErrors, FieldError{
			Field:   "seats",
			Message: "количество мест должно быть больше нуля",
		})
	}

	if len(validationErrors) > 0 {
		return ValidationWithFields(validationErrors...)
	}

	return nil
}
