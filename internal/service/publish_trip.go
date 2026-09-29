package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"share_trip/internal/domain"
	"share_trip/internal/observability/logctx"
	"share_trip/internal/outbox"
	"share_trip/internal/storage/repository"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type PublishTripRequest struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type PublishTripResponse struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
	Status        domain.TripStatus
	CreatedAt     time.Time
}

func (s *TripService) PublishTrip(ctx context.Context, req PublishTripRequest) (*PublishTripResponse, error) {
	tracer := otel.Tracer("TripService")
	ctx, span := tracer.Start(ctx, "TripService.PublishTrip")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "publish_trip"),
		attribute.String("trip_id", req.TripID.String()),
		attribute.String("driver_id", req.DriverID.String()),
	)

	started := time.Now()
	publishEventCreated := false

	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "PublishTrip"),
	)

	logger.InfoContext(ctx, "публикация поездки в service начата")

	if err := validatePublishTripRequest(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.WarnContext(
			ctx,
			"публикация поездки не выполнена: ошибка валидации",
			slog.Any("error", err),
		)
		return nil, err
	}

	var publishedTrip domain.Trip
	var fromStatus domain.TripStatus

	err := s.runTripTx(ctx, func(ctx context.Context, tx TripRepositoryTx) error {
		trip, err := tx.GetForUpdateByID(ctx, req.TripID)
		if err != nil {
			return publishTripError(req.TripID, err)
		}

		fromStatus = trip.Status
		if err := trip.Publish(req.DriverID); err != nil {
			return publishTripError(req.TripID, err)
		}

		if fromStatus == trip.Status {
			publishedTrip = trip
			return nil
		}

		updatedTrip, err := tx.UpdateStatus(ctx, trip.ID, trip.Status)
		if err != nil {
			return publishTripError(req.TripID, err)
		}
		publishedTrip = updatedTrip

		history := domain.TripHistory{
			ID:         uuid.New(),
			TripID:     trip.ID,
			FromStatus: &fromStatus,
			ToStatus:   trip.Status,
			CreatedAt:  time.Now(),
		}
		if err := tx.CreateHistory(ctx, history); err != nil {
			return publishTripError(req.TripID, err)
		}

		event, err := outbox.NewTripPublishedEvent(trip.ID, trip.DriverID)
		if err != nil {
			return publishTripError(req.TripID, err)
		}
		if err := tx.CreateOutboxEvent(ctx, event); err != nil {
			return publishTripError(req.TripID, err)
		}
		publishEventCreated = true
		return nil
	})

	span.SetAttributes(
		attribute.String("from_status", string(fromStatus)),
		attribute.String("to_status", string(publishedTrip.Status)),
	)

	if err != nil {
		tripErr := publishTripError(req.TripID, err)
		span.RecordError(tripErr)
		span.SetStatus(codes.Error, tripErr.Error())
		logger.ErrorContext(
			ctx,
			"публикация поездки в service не выполнена",
			slog.Any("error", err),
		)
		return nil, tripErr
	}

	if publishEventCreated {
		s.metrics.TripPublishTotal.WithLabelValues(metricResultSuccess).Inc()
		s.metrics.TripPublishDuration.WithLabelValues(metricResultSuccess).
			Observe(time.Since(started).Seconds())
	}

	publishTripResult := toPublishTripResult(publishedTrip)
	logger.InfoContext(
		ctx,
		"публикация поездки в service завершена",
		slog.String("status", string(publishTripResult.Status)),
	)
	span.SetAttributes(attribute.String("status", string(publishTripResult.Status)))
	return &publishTripResult, nil
}

func validatePublishTripRequest(req PublishTripRequest) error {
	var validationErrors []FieldError

	if req.TripID == uuid.Nil {
		validationErrors = append(validationErrors, FieldError{
			Field:   "tripId",
			Message: "ID поездки обязателен",
		})
	}
	if req.DriverID == uuid.Nil {
		validationErrors = append(validationErrors, FieldError{
			Field:   "driverId",
			Message: "ID водителя обязателен",
		})
	}

	if len(validationErrors) > 0 {
		return ValidationWithFields(validationErrors...)
	}

	return nil
}

func toPublishTripResult(trip domain.Trip) PublishTripResponse {
	return PublishTripResponse{
		ID:            trip.ID,
		DriverID:      trip.DriverID,
		FromPoint:     trip.FromPoint,
		ToPoint:       trip.ToPoint,
		DepartureTime: trip.DepartureTime,
		Seats:         trip.Seats,
		Status:        trip.Status,
		CreatedAt:     trip.CreatedAt,
	}
}

func publishTripError(tripID uuid.UUID, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	if errors.Is(err, repository.ErrNotFound) {
		return NotFound(fmt.Sprintf("поездка не найдена: %s", tripID))
	}
	if errors.Is(err, domain.ErrTripDriverMismatch) {
		return Forbidden(domain.ErrTripDriverMismatch.Error())
	}
	if errors.Is(err, domain.ErrTripInvalidStatusTransition) {
		return Conflict(err.Error())
	}

	return err
}
