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
		span.SetStatus(codes.Error, err.Error())
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

	checkResult, err := s.contractChecker.CheckService(ctx, trip.DriverID.String(), "tripCreation")
	if err != nil {
		logger.ErrorContext(ctx, "не удалось проверить права через Contract Service", slog.Any("error", err))
		return nil, fmt.Errorf("не удалось проверить права на старт поездки: %w", err)
	}

	if !checkResult.Allowed {
		logger.WarnContext(ctx, "старт поездки запрещен по контракту", slog.String("reason", checkResult.Reason))
		return nil, Forbidden(fmt.Sprintf("старт поездки запрещен: %s", checkResult.Reason))
	}

	var startedTrip domain.Trip

	err = s.runTripTx(ctx, func(ctx context.Context, tripRepositoryTx TripRepositoryTx) error {
		tripTx, err := tripRepositoryTx.GetForUpdateByID(ctx, req.TripID)
		if err != nil {
			return startTripError(req.TripID, err)
		}

		fromStatus := tripTx.Status

		if err := tripTx.Start(req.DriverID); err != nil {
			return startTripError(req.TripID, err)
		}

		startedTrip = tripTx

		if fromStatus != tripTx.Status {
			_, err = tripRepositoryTx.UpdateStatus(ctx, tripTx.ID, tripTx.Status)
			if err != nil {
				return startTripError(req.TripID, err)
			}

			history := domain.TripHistory{
				ID:         uuid.New(),
				TripID:     tripTx.ID,
				FromStatus: &fromStatus,
				ToStatus:   tripTx.Status,
				CreatedAt:  time.Now(),
			}
			if err := tripRepositoryTx.CreateHistory(ctx, history); err != nil {
				return startTripError(req.TripID, err)
			}

			event, err := outbox.NewTripStartedEvent(tripTx.ID)
			if err != nil {
				return startTripError(req.TripID, err)
			}
			if err := tripRepositoryTx.CreateOutboxEvent(ctx, event); err != nil {
				return startTripError(req.TripID, err)
			}
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
