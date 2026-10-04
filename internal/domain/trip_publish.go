package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PublishTripRequest struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type PublishTripResponse struct {
	Trip          Trip
	FromStatus    TripStatus
	StatusChanged bool
}

func PublishTrip(ctx context.Context, req PublishTripRequest, tx TripRepositoryTx) (PublishTripResponse, error) {
	trip, err := tx.GetForUpdateByID(ctx, req.TripID)
	if err != nil {
		return PublishTripResponse{}, err
	}

	fromStatus := trip.Status

	if trip.DriverID != req.DriverID {
		return PublishTripResponse{}, ErrTripDriverMismatch
	}

	if trip.Status == TripStatusPublished {
		return PublishTripResponse{Trip: trip, FromStatus: fromStatus, StatusChanged: false}, nil
	}

	if !canTransitionTripStatus(fromStatus, TripStatusPublished) {
		return PublishTripResponse{}, fmt.Errorf("%w: %s -> %s", ErrTripInvalidStatusTransition, fromStatus, TripStatusPublished)
	}

	trip.Status = TripStatusPublished

	updatedTrip, err := tx.UpdateStatus(ctx, trip.ID, trip.Status)
	if err != nil {
		return PublishTripResponse{}, err
	}

	history := TripHistory{
		ID:         uuid.New(),
		TripID:     trip.ID,
		FromStatus: &fromStatus,
		ToStatus:   trip.Status,
		CreatedAt:  time.Now(),
	}

	if err := tx.CreateHistory(ctx, history); err != nil {
		return PublishTripResponse{}, err
	}

	return PublishTripResponse{
		Trip:          updatedTrip,
		FromStatus:    fromStatus,
		StatusChanged: true,
	}, nil
}
