package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"share_trip/internal/domain"
	"share_trip/internal/outbox"
	"share_trip/internal/storage/repository"

	"github.com/google/uuid"
)

type PublishTripCommand struct {
	TripID   uuid.UUID
	DriverID uuid.UUID
}

func (s *TripService) PublishTrip(ctx context.Context, cmd PublishTripCommand) error {
	if err := validatePublishTripCommand(cmd); err != nil {
		return err
	}

	err := s.runTripTx(ctx, func(ctx context.Context, trips TripRepositoryTx) error {
		trip, err := trips.GetForUpdateByID(ctx, cmd.TripID)
		if err != nil {
			return publishTripError(cmd.TripID, err)
		}

		fromStatus := trip.Status
		if err := trip.Publish(cmd.DriverID); err != nil {
			return publishTripError(cmd.TripID, err)
		}

		if fromStatus == trip.Status {
			return nil
		}

		if err := trips.UpdateStatus(ctx, trip.ID, trip.Status); err != nil {
			return publishTripError(cmd.TripID, err)
		}

		history := domain.TripHistory{
			ID:         uuid.New(),
			TripID:     trip.ID,
			FromStatus: &fromStatus,
			ToStatus:   trip.Status,
			CreatedAt:  time.Now(),
		}
		if err := trips.AppendHistory(ctx, history); err != nil {
			return publishTripError(cmd.TripID, err)
		}

		event, err := outbox.NewTripPublishedEvent(trip.ID)
		if err != nil {
			return publishTripError(cmd.TripID, err)
		}
		if err := trips.AppendOutboxEvent(ctx, event); err != nil {
			return publishTripError(cmd.TripID, err)
		}

		return nil
	})
	if err != nil {
		return publishTripError(cmd.TripID, err)
	}

	return nil
}

func validatePublishTripCommand(cmd PublishTripCommand) error {
	var validationErrors []FieldError

	if cmd.TripID == uuid.Nil {
		validationErrors = append(validationErrors, FieldError{
			Field:   "tripId",
			Message: "ID поездки обязателен",
		})
	}
	if cmd.DriverID == uuid.Nil {
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

func publishTripError(tripID uuid.UUID, err error) error {
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
		return Unprocessable(err.Error())
	}

	return err
}
