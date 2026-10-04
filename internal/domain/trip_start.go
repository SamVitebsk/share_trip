package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StartTripRequest struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

type StartTripResponse struct {
	Trip          Trip
	FromStatus    TripStatus
	StatusChanged bool
}

func StartTrip(ctx context.Context, req StartTripRequest, tx TripRepositoryTx) (StartTripResponse, error) {
	trip, err := tx.GetForUpdateByID(ctx, req.TripID)
	if err != nil {
		return StartTripResponse{}, err
	}

	fromStatus := trip.Status

	if trip.DriverID != req.DriverID {
		return StartTripResponse{}, ErrTripDriverMismatch
	}

	if trip.Status == TripStatusStarted {
		return StartTripResponse{Trip: trip, FromStatus: fromStatus, StatusChanged: false}, nil
	}

	if !canTransitionTripStatus(fromStatus, TripStatusStarted) {
		return StartTripResponse{}, fmt.Errorf("%w: %s -> %s", ErrTripInvalidStatusTransition, fromStatus, TripStatusStarted)
	}

	trip.Status = TripStatusStarted

	updatedTrip, err := tx.UpdateStatus(ctx, trip.ID, trip.Status)
	if err != nil {
		return StartTripResponse{}, err
	}

	history := TripHistory{
		ID:         uuid.New(),
		TripID:     trip.ID,
		FromStatus: &fromStatus,
		ToStatus:   trip.Status,
		CreatedAt:  time.Now(),
	}

	if err := tx.CreateHistory(ctx, history); err != nil {
		return StartTripResponse{}, err
	}

	return StartTripResponse{
		Trip:          updatedTrip,
		FromStatus:    fromStatus,
		StatusChanged: true,
	}, nil
}
