package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"share_trip/internal/domain"
	"share_trip/internal/observability/logctx"
	"share_trip/internal/outbox"
	"share_trip/internal/storage/repository"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type StartTripRequest struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type StartTripResponse struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
	Status        domain.TripStatus
	CreatedAt     time.Time
}

func (s *TripService) StartTrip(ctx context.Context, req StartTripRequest) (*StartTripResponse, error) {
	tracer := otel.Tracer("TripService")
	ctx, span := tracer.Start(ctx, "TripService.StartTrip")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "start_trip"),
		attribute.String("trip_id", req.TripID.String()),
		attribute.String("driver_id", req.DriverID.String()),
	)

	if err := validateStartTripRequest(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "trip start failed")
		return nil, err
	}

	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "StartTrip"),
	)

	trip, err := s.tripRepository.GetByID(ctx, req.TripID)
	if err != nil {
		return nil, startTripError(req.TripID, err)
	}

	if trip.DriverID != req.DriverID {
		return nil, startTripError(req.TripID, domain.ErrTripDriverMismatch)
	}

	checkServiceRequest := CheckServiceRequest{DriverID: trip.DriverID.String(), ServiceCode: "tripCreation"}
	checkResult, err := s.contractChecker.CheckService(ctx, checkServiceRequest)
	if err != nil {
		logger.ErrorContext(ctx, "не удалось проверить права через Contract Service", slog.Any("error", err))
		return nil, fmt.Errorf("не удалось проверить права на старт поездки: %w", err)
	}

	if !checkResult.Allowed {
		logger.WarnContext(ctx, "старт поездки запрещен по контракту", slog.String("reason", checkResult.Reason))
		return nil, Forbidden(fmt.Sprintf("старт поездки запрещен: %s", checkResult.Reason))
	}

	var startedTrip domain.Trip

	err = s.runTripTx(ctx, func(ctx context.Context, tx TripRepositoryTx) error {
		res, err := domain.StartTrip(ctx, domain.StartTripRequest{
			TripID:   req.TripID,
			DriverID: req.DriverID,
		}, tx)

		if err != nil {
			return startTripError(req.TripID, err)
		}

		startedTrip = res.Trip

		if res.StatusChanged {
			outboxReq := outbox.TripEventRequest{
				TripID:   res.Trip.ID,
				DriverID: res.Trip.DriverID,
			}
			event, err := outbox.NewTripStartedEvent(ctx, outboxReq)
			if err != nil {
				return startTripError(req.TripID, err)
			}
			if err := tx.CreateOutboxEvent(ctx, event); err != nil {
				return startTripError(req.TripID, err)
			}
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "trip start failed")
		return nil, startTripError(req.TripID, err)
	}

	return toStartTripResponse(startedTrip), nil
}

func toStartTripResponse(trip domain.Trip) *StartTripResponse {
	return &StartTripResponse{
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

func startTripError(tripID uuid.UUID, err error) error {
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

func validateStartTripRequest(req StartTripRequest) error {
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
