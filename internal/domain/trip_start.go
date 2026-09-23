package domain

import (
	"fmt"

	"github.com/google/uuid"
)

func (t *Trip) Start(driverID uuid.UUID) error {
	if t.DriverID != driverID {
		return ErrTripDriverMismatch
	}

	if t.Status == TripStatusStarted {
		return nil
	}

	fromStatus := t.Status
	if fromStatus != TripStatusPublished {
		return fmt.Errorf("%w: %s -> %s", ErrTripInvalidStatusTransition, fromStatus, TripStatusStarted)
	}

	t.Status = TripStatusStarted

	return nil
}
