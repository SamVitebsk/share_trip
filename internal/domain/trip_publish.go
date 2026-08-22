package domain

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrTripDriverMismatch          = errors.New("водитель не является владельцем поездки")
	ErrTripInvalidStatusTransition = errors.New("недопустимый переход статуса поездки")
)

func (t *Trip) Publish(driverID uuid.UUID) error {
	if t.DriverID != driverID {
		return ErrTripDriverMismatch
	}

	if t.Status == TripStatusPublished {
		return nil
	}

	fromStatus := t.Status
	if !CanTransitionTripStatus(fromStatus, TripStatusPublished) {
		return fmt.Errorf("%w: %s -> %s", ErrTripInvalidStatusTransition, fromStatus, TripStatusPublished)
	}

	t.Status = TripStatusPublished

	return nil
}
