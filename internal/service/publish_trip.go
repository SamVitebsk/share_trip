package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"share_trip/internal/domain"
	"share_trip/internal/observability/logctx"
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
	logger := logctx.Logger(ctx).With(
		slog.String("service", "TripService"),
		slog.String("operation", "PublishTrip"),
	)

	logger.InfoContext(ctx, "публикация поездки в service начата")

	if err := validatePublishTripCommand(cmd); err != nil {
		logger.WarnContext(
			ctx,
			"публикация поездки не выполнена: ошибка валидации",
			slog.Any("error", err),
		)
		return nil, err
	}

	var publishedTrip domain.Trip
	err := s.runTripTx(ctx, func(ctx context.Context, tripRepositoryTx TripRepositoryTx) error {
		logger.InfoContext(ctx, "получение поездки для публикации начато")

		trip, err := tripRepositoryTx.GetForUpdateByID(ctx, cmd.TripID)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"получение поездки для публикации не выполнено",
				slog.Any("error", err),
			)
			return publishTripError(cmd.TripID, err)
		}

		fromStatus := trip.Status
		logger.InfoContext(
			ctx,
			"проверка доменных правил публикации начата",
			slog.String("from_status", string(fromStatus)),
		)

		if err := trip.Publish(cmd.DriverID); err != nil {
			logger.WarnContext(
				ctx,
				"публикация поездки не выполнена: доменное правило не пройдено",
				slog.String("from_status", string(fromStatus)),
				slog.Any("error", err),
			)
			return publishTripError(cmd.TripID, err)
		}

		if fromStatus == trip.Status {
			logger.InfoContext(
				ctx,
				"публикация поездки не требует изменения статуса",
				slog.String("status", string(trip.Status)),
			)
			publishedTrip = trip
			return nil
		}

		logger.InfoContext(
			ctx,
			"статус поездки изменен доменной логикой",
			slog.String("from_status", string(fromStatus)),
			slog.String("to_status", string(trip.Status)),
		)

		updatedTrip, err := tripRepositoryTx.UpdateStatus(ctx, trip.ID, trip.Status)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"обновление статуса поездки не выполнено",
				slog.Any("error", err),
			)
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
			logger.ErrorContext(
				ctx,
				"создание истории публикации поездки не выполнено",
				slog.Any("error", err),
			)
			return publishTripError(cmd.TripID, err)
		}

		event, err := outbox.NewTripPublishedEvent(trip.ID)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"создание outbox-события публикации поездки не выполнено",
				slog.Any("error", err),
			)
			return publishTripError(cmd.TripID, err)
		}
		if err := tripRepositoryTx.CreateOutboxEvent(ctx, event); err != nil {
			logger.ErrorContext(
				ctx,
				"сохранение outbox-события публикации поездки не выполнено",
				slog.Any("error", err),
			)
			return publishTripError(cmd.TripID, err)
		}

		return nil
	})
	if err != nil {
		logger.ErrorContext(
			ctx,
			"публикация поездки в service не выполнена",
			slog.Any("error", err),
		)
		return nil, publishTripError(cmd.TripID, err)
	}

	publishTripResult := toPublishTripResult(publishedTrip)

	logger.InfoContext(
		ctx,
		"публикация поездки в service завершена",
		slog.String("status", string(publishTripResult.Status)),
	)

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
