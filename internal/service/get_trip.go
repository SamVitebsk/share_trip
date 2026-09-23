package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"share_trip/internal/domain"
	"share_trip/internal/observability/logctx"
	"share_trip/internal/storage/repository"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type GetTripRequest struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type TripResponse struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
	Status        domain.TripStatus
	CreatedAt     time.Time
}

func (s *TripService) GetTrip(ctx context.Context, req GetTripRequest) (TripResponse, error) {
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "GetTrip"),
	)

	tracer := otel.Tracer("TripUsecase")
	ctx, span := tracer.Start(ctx, "TripUsecase.GetTrip")
	defer span.End()
	span.SetAttributes(
		attribute.String("operation", "get_trip"),
		attribute.String("trip_id", req.TripID.String()),
		attribute.String("driver_id", req.DriverID.String()),
	)

	logger.InfoContext(ctx, "получение поездки в service начато")

	trip, err := s.tripRepository.GetByID(ctx, req.TripID)
	if errors.Is(err, repository.ErrNotFound) {
		tripErr := NotFound(fmt.Sprintf("поездка не найдена: %s", req.TripID))
		span.RecordError(tripErr)
		span.SetStatus(codes.Error, tripErr.Error())
		logger.WarnContext(
			ctx,
			"получение поездки не выполнено: поездка не найдена",
			slog.Any("error", err),
		)
		return TripResponse{}, tripErr
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.ErrorContext(
			ctx,
			"получение поездки не выполнено: ошибка repository",
			slog.Any("error", err),
		)
		return TripResponse{}, err
	}

	if trip.DriverID != req.DriverID {
		tripErr := Forbidden("доступ к поездке запрещен")
		span.RecordError(tripErr)
		span.SetStatus(codes.Error, tripErr.Error())
		logger.WarnContext(
			ctx,
			"получение поездки не выполнено: поездка принадлежит другому водителю",
			slog.String("driver_id", trip.DriverID.String()),
			slog.String("request_driver_id", req.DriverID.String()),
		)
		return TripResponse{}, tripErr
	}

	tripView := toTripResponse(trip)

	logger.InfoContext(
		ctx,
		"получение поездки в service завершено",
		slog.String("driver_id", tripView.DriverID.String()),
		slog.String("status", string(tripView.Status)),
	)
	span.SetAttributes(
		attribute.String("driver_id", tripView.DriverID.String()),
		attribute.String("status", string(tripView.Status)),
	)

	return tripView, nil
}

func toTripResponse(trip domain.Trip) TripResponse {
	return TripResponse{
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
