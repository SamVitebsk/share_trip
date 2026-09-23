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

type StartTripResponse struct {
	ID            string `json:"id"`
	DriverID      string `json:"driverId"`
	FromPoint     string `json:"fromPoint"`
	ToPoint       string `json:"toPoint"`
	DepartureTime string `json:"departureTime"`
	Seats         int    `json:"seats"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
}

func (h *TripHandler) StartTrip(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger := logctx.Logger(ctx).With(
		slog.String("handler", "StartTrip"),
	)

	tracer := otel.Tracer("trip-api")
	ctx, span := tracer.Start(ctx, "StartTripHandler")
	defer span.End()

	c.SetUserContext(ctx)
	c.Set("trace-id", span.SpanContext().TraceID().String())
	span.SetAttributes(attribute.String("operation", "start_trip"))

	driverID, err := driverIDFromClaims(c)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn(
			"старт поездки не выполнен: данные пользователя не получены",
			slog.Any("error", err),
		)
		return err
	}

	startTripRequest, err := parseStartTripRequest(c.Params("tripId"), driverID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn(
			"старт поездки не выполнен: некорректный запрос",
			slog.Any("error", err),
		)
		return writeError(c, ErrorCodeValidation, err.Error())
	}

	logger = logger.With(
		slog.String("trip_id", startTripRequest.TripID.String()),
		slog.String("driver_id", startTripRequest.DriverID.String()),
	)
	span.SetAttributes(
		attribute.String("trip_id", startTripRequest.TripID.String()),
		attribute.String("driver_id", startTripRequest.DriverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	c.SetUserContext(ctx)

	logger.Info("запрос на старт поездки принят")

	startTripResponse, err := h.tripService.StartTrip(ctx, *startTripRequest)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error(
			"старт поездки не выполнен",
			slog.Any("error", err),
		)
		return writeServiceError(c, err)
	}

	logger.Info(
		"старт поездки завершен успешно",
		slog.String("status", string(startTripResponse.Status)),
	)
	span.SetAttributes(attribute.String("status", string(startTripResponse.Status)))

	return c.Status(fiber.StatusOK).JSON(toStartTripResponse(startTripResponse))
}

func parseStartTripRequest(tripIDParam string, driverID uuid.UUID) (*service.StartTripRequest, error) {
	tripID, err := uuid.Parse(tripIDParam)
	if err != nil {
		return nil, errInvalidTripID
	}

	return &service.StartTripRequest{
		TripID:   tripID,
		DriverID: driverID,
	}, nil
}

func toStartTripResponse(result *service.StartTripResponse) StartTripResponse {
	return StartTripResponse{
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
