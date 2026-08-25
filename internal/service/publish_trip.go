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

type PublishTripResult struct {
	ID            uuid.UUID
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
	Status        domain.TripStatus
	CreatedAt     time.Time
}

func (s *TripService) PublishTrip(ctx context.Context, cmd PublishTripCommand) (*PublishTripResult, error) {
	if err := validatePublishTripCommand(cmd); err != nil {
		return nil, err
	}

	var publishedTrip domain.Trip
	err := s.runTripTx(ctx, func(ctx context.Context, tripRepositoryTx TripRepositoryTx) error {
		trip, err := tripRepositoryTx.GetForUpdateByID(ctx, cmd.TripID)
		if err != nil {
			return publishTripError(cmd.TripID, err)
		}

		fromStatus := trip.Status
		if err := trip.Publish(cmd.DriverID); err != nil {
			return publishTripError(cmd.TripID, err)
		}

		if fromStatus == trip.Status {
			publishedTrip = trip
			return nil
		}

		updatedTrip, err := tripRepositoryTx.UpdateStatus(ctx, trip.ID, trip.Status)
		if err != nil {
			return publishTripError(cmd.TripID, err)
		}
		publishedTrip = updatedTrip

		history := domain.TripHistory{
			ID:         uuid.New(),
			TripID:     trip.ID,
			FromStatus: &fromStatus,
			ToStatus:   trip.Status,
			CreatedAt:  time.Now(),
		}
		if err := tripRepositoryTx.CreateHistory(ctx, history); err != nil {
			return publishTripError(cmd.TripID, err)
		}

		event, err := outbox.NewTripPublishedEvent(trip.ID)
		if err != nil {
			return publishTripError(cmd.TripID, err)
		}
		if err := tripRepositoryTx.CreateOutboxEvent(ctx, event); err != nil {
			return publishTripError(cmd.TripID, err)
		}

		return nil
	})
	if err != nil {
		return nil, publishTripError(cmd.TripID, err)
	}

	publishTripResult := toPublishTripResult(publishedTrip)

	return &publishTripResult, nil
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

func toPublishTripResult(trip domain.Trip) PublishTripResult {
	return PublishTripResult{
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
		return Conflict(err.Error())
	}

	return err
}
