package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"share_trip/internal/domain"
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
	trip, err := s.tripRepository.GetByID(ctx, query.TripID)
	if errors.Is(err, repository.ErrNotFound) {
		return TripView{}, NotFound(fmt.Sprintf("поездка не найдена: %s", query.TripID))
	}
	if err != nil {
		return TripView{}, err
	}

	return tripViewFromDomain(trip), nil
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
