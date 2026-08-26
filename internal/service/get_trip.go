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
)

type GetTripQuery struct {
	TripID uuid.UUID
}

type TripView struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
	Status        domain.TripStatus
	CreatedAt     time.Time
}

func (s *TripService) GetTrip(ctx context.Context, query GetTripQuery) (TripView, error) {
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "GetTrip"),
	)

	logger.InfoContext(ctx, "получение поездки в service начато")

	trip, err := s.tripRepository.GetByID(ctx, query.TripID)
	if errors.Is(err, repository.ErrNotFound) {
		logger.WarnContext(
			ctx,
			"получение поездки не выполнено: поездка не найдена",
			slog.Any("error", err),
		)
		return TripView{}, NotFound(fmt.Sprintf("поездка не найдена: %s", query.TripID))
	}
	if err != nil {
		logger.ErrorContext(
			ctx,
			"получение поездки не выполнено: ошибка repository",
			slog.Any("error", err),
		)
		return TripView{}, err
	}

	tripView := tripViewFromDomain(trip)

	logger.InfoContext(
		ctx,
		"получение поездки в service завершено",
		slog.String("driver_id", tripView.DriverID.String()),
		slog.String("status", string(tripView.Status)),
	)

	return tripView, nil
}

func tripViewFromDomain(trip domain.Trip) TripView {
	return TripView{
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
