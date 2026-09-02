package api

import (
	"log/slog"

	"share_trip/internal/observability/logctx"
	"share_trip/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type PublishTripResponse struct {
	ID            string `json:"id"`
	DriverID      string `json:"driverId"`
	FromPoint     string `json:"fromPoint"`
	ToPoint       string `json:"toPoint"`
	DepartureTime string `json:"departureTime"`
	Seats         int    `json:"seats"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
}

func (h *TripHandler) PublishTrip(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger := logctx.Logger(ctx).With(
		slog.String("handler", "PublishTrip"),
	)

	tracer := otel.Tracer("trip-api")
	ctx, span := tracer.Start(ctx, "PublishTripHandler")
	defer span.End()

	c.SetUserContext(ctx)
	c.Set("trace-id", span.SpanContext().TraceID().String())
	span.SetAttributes(attribute.String("operation", "publish_trip"))

	driverID, err := driverIDFromClaims(c)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn(
			"публикация поездки не выполнена: данные пользователя не получены",
			slog.Any("error", err),
		)
		return err
	}

	publishTripCommand, err := parsePublishTripCommandFromRequest(c.Params("tripId"), driverID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn(
			"публикация поездки не выполнена: некорректный запрос",
			slog.Any("error", err),
		)
		return writeError(c, ErrorCodeValidation, err.Error())
	}

	logger = logger.With(
		slog.String("trip_id", publishTripCommand.TripID.String()),
		slog.String("driver_id", publishTripCommand.DriverID.String()),
	)
	span.SetAttributes(
		attribute.String("trip_id", publishTripCommand.TripID.String()),
		attribute.String("driver_id", publishTripCommand.DriverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	c.SetUserContext(ctx)

	logger.Info("запрос на публикацию поездки принят")

	result, err := h.tripService.PublishTrip(ctx, *publishTripCommand)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error(
			"публикация поездки не выполнена",
			slog.Any("error", err),
		)
		return writeServiceError(c, err)
	}

	logger.Info(
		"публикация поездки завершена",
		slog.String("status", string(result.Status)),
	)
	span.SetAttributes(attribute.String("status", string(result.Status)))

	return c.Status(fiber.StatusOK).JSON(toPublishTripResponse(result))
}

func parsePublishTripCommandFromRequest(tripIDParam string, driverID uuid.UUID) (*service.PublishTripCommand, error) {
	tripID, err := uuid.Parse(tripIDParam)
	if err != nil {
		return nil, errInvalidTripID
	}

	return &service.PublishTripCommand{
		TripID:   tripID,
		DriverID: driverID,
	}, nil
}

func toPublishTripResponse(result *service.PublishTripResult) PublishTripResponse {
	return PublishTripResponse{
		ID:            result.ID.String(),
		DriverID:      result.DriverID.String(),
		FromPoint:     result.FromPoint,
		ToPoint:       result.ToPoint,
		DepartureTime: formatHTTPTime(result.DepartureTime),
		Seats:         result.Seats,
		Status:        string(result.Status),
		CreatedAt:     formatHTTPTime(result.CreatedAt),
	}
}
