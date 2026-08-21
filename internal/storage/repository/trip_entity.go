package repository

import (
	"share_trip/internal/domain"
	"time"

	"github.com/google/uuid"
)

type tripEntity struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
	Status        string
	CreatedAt     time.Time
}

type tripHistoryEntity struct {
	ID         uuid.UUID
	TripID     uuid.UUID
	FromStatus *string
	ToStatus   string
	CreatedAt  time.Time
}

func toTripEntity(trip domain.Trip) tripEntity {
	return tripEntity{
		ID:            trip.ID,
		DriverID:      trip.DriverID,
		FromPoint:     trip.FromPoint,
		ToPoint:       trip.ToPoint,
		DepartureTime: trip.DepartureTime,
		Seats:         trip.Seats,
		Status:        string(trip.Status),
		CreatedAt:     trip.CreatedAt,
	}
}

func toDomainTrip(entity tripEntity) domain.Trip {
	return domain.Trip{
		ID:            entity.ID,
		DriverID:      entity.DriverID,
		FromPoint:     entity.FromPoint,
		ToPoint:       entity.ToPoint,
		DepartureTime: entity.DepartureTime,
		Seats:         entity.Seats,
		Status:        domain.TripStatus(entity.Status),
		CreatedAt:     entity.CreatedAt,
	}
}

func toTripHistoryEntity(history domain.TripHistory) tripHistoryEntity {
	return tripHistoryEntity{
		ID:         history.ID,
		TripID:     history.TripID,
		FromStatus: tripStatusPointerToStringPointer(history.FromStatus),
		ToStatus:   string(history.ToStatus),
		CreatedAt:  history.CreatedAt,
	}
}

func tripStatusPointerToStringPointer(status *domain.TripStatus) *string {
	if status == nil {
		return nil
	}

	statusStr := string(*status)
	return &statusStr
}
