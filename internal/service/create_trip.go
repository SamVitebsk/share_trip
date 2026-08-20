package service

import (
	"context"
	"strings"
	"time"

	"share_trip/internal/domain"

	"github.com/google/uuid"
)

type CreateTripCommand struct {
	DriverID      uuid.UUID
	FromPoint     string
	ToPoint       string
	DepartureTime time.Time
	Seats         int
}

type CreateTripResult struct {
	TripID uuid.UUID
}

func (s *TripService) CreateTrip(ctx context.Context, cmd CreateTripCommand) (CreateTripResult, error) {
	now := time.Now()
	cmd = normalizeCreateTripCommand(cmd)
	if err := validateCreateTripCommand(cmd, now); err != nil {
		return CreateTripResult{}, err
	}

	trip := domain.Trip{
		ID:            uuid.New(),
		DriverID:      cmd.DriverID,
		FromPoint:     cmd.FromPoint,
		ToPoint:       cmd.ToPoint,
		DepartureTime: cmd.DepartureTime,
		Seats:         cmd.Seats,
		Status:        domain.TripStatusDraft,
		CreatedAt:     now,
	}
	history := domain.TripHistory{
		ID:         uuid.New(),
		TripID:     trip.ID,
		FromStatus: nil,
		ToStatus:   trip.Status,
		CreatedAt:  now,
	}

	err := s.trips.Create(ctx, trip, history)
	if err != nil {
		return CreateTripResult{}, err
	}

	return CreateTripResult{TripID: trip.ID}, nil
}

func normalizeCreateTripCommand(cmd CreateTripCommand) CreateTripCommand {
	cmd.FromPoint = strings.TrimSpace(cmd.FromPoint)
	cmd.ToPoint = strings.TrimSpace(cmd.ToPoint)
	return cmd
}

func validateCreateTripCommand(cmd CreateTripCommand, now time.Time) error {
	var validationErrors []FieldError

	if cmd.DriverID == uuid.Nil {
		validationErrors = append(validationErrors, FieldError{
			Field:   "driverId",
			Message: "ID водителя обязателен",
		})
	}
	if cmd.FromPoint == "" {
		validationErrors = append(validationErrors, FieldError{
			Field:   "fromPoint",
			Message: "пункт отправления обязателен",
		})
	}
	if cmd.ToPoint == "" {
		validationErrors = append(validationErrors, FieldError{
			Field:   "toPoint",
			Message: "пункт назначения обязателен",
		})
	}
	if cmd.FromPoint != "" && cmd.ToPoint != "" && cmd.FromPoint == cmd.ToPoint {
		validationErrors = append(validationErrors, FieldError{
			Field:   "route",
			Message: "пункт отправления и пункт назначения должны отличаться",
		})
	}
	if cmd.DepartureTime.IsZero() {
		validationErrors = append(validationErrors, FieldError{
			Field:   "departureTime",
			Message: "время отправления обязательно",
		})
	} else if !cmd.DepartureTime.After(now) {
		validationErrors = append(validationErrors, FieldError{
			Field:   "departureTime",
			Message: "время отправления должно быть в будущем",
		})
	}
	if cmd.Seats <= 0 {
		validationErrors = append(validationErrors, FieldError{
			Field:   "seats",
			Message: "количество мест должно быть больше нуля",
		})
	}

	if len(validationErrors) > 0 {
		return ValidationWithFields(validationErrors...)
	}

	return nil
}
