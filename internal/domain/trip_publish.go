package domain

import (
	"fmt"

	"github.com/google/uuid"
)

func (t *Trip) Publish(driverID uuid.UUID) error {
	if t.DriverID != driverID {
		return ErrTripDriverMismatch
	}

	if t.Status == TripStatusPublished {
		return nil
	}

	fromStatus := t.Status
	if !canTransitionTripStatus(fromStatus, TripStatusPublished) {
		return fmt.Errorf("%w: %s -> %s", ErrTripInvalidStatusTransition, fromStatus, TripStatusPublished)
	}

	t.Status = TripStatusPublished

	return nil
}
