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

type CreateTripRequest struct {
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
}

type CreateTripResponse struct {
	TripID uuid.UUID
}

func (s *TripService) CreateTrip(ctx context.Context, req CreateTripRequest) (CreateTripResponse, error) {
	tracer := otel.Tracer("TripService")
	ctx, span := tracer.Start(ctx, "TripService.CreateTrip")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "create_trip"),
		attribute.String("driver_id", req.DriverID.String()),
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
	req = normalizeCreateTripRequest(req)
	if err := validateCreateTripRequest(req, now); err != nil {
		result = metricResultFromError(err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn(
			"создание поездки не выполнено: ошибка валидации",
			slog.Any("error", err),
		)
		return CreateTripResponse{}, err
	}

	trip := domain.Trip{
		ID:            uuid.New(),
		DriverID:      req.DriverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		DepartureTime: req.DepartureTime,
		Seats:         req.Seats,
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
		return CreateTripResponse{}, err
	}

	logger.Info(
		"создание поездки в service завершено",
		slog.String("trip_id", trip.ID.String()),
	)

	return CreateTripResponse{TripID: trip.ID}, nil
}

func normalizeCreateTripRequest(req CreateTripRequest) CreateTripRequest {
	req.FromPoint = strings.TrimSpace(req.FromPoint)
	req.ToPoint = strings.TrimSpace(req.ToPoint)
	return req
}

func validateCreateTripRequest(req CreateTripRequest, now time.Time) error {
	var validationErrors []FieldError

	if req.DriverID == uuid.Nil {
		validationErrors = append(validationErrors, FieldError{
			Field:   "driverId",
			Message: "ID водителя обязателен",
		})
	}
	if req.FromPoint == "" {
		validationErrors = append(validationErrors, FieldError{
			Field:   "fromPoint",
			Message: "пункт отправления обязателен",
		})
	}
	if req.ToPoint == "" {
		validationErrors = append(validationErrors, FieldError{
			Field:   "toPoint",
			Message: "пункт назначения обязателен",
		})
	}
	if req.FromPoint != "" && req.ToPoint != "" && req.FromPoint == req.ToPoint {
		validationErrors = append(validationErrors, FieldError{
			Field:   "route",
			Message: "пункт отправления и пункт назначения должны отличаться",
		})
	}
	if req.DepartureTime.IsZero() {
		validationErrors = append(validationErrors, FieldError{
			Field:   "departureTime",
			Message: "время отправления обязательно",
		})
	} else if !req.DepartureTime.After(now) {
		validationErrors = append(validationErrors, FieldError{
			Field:   "departureTime",
			Message: "время отправления должно быть в будущем",
		})
	}
	if req.Seats <= 0 {
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
