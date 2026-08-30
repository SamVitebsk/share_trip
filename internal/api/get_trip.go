package api

import (
	"errors"
	"log/slog"
	"share_trip/internal/observability/logctx"
	"share_trip/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var (
	errInvalidTripID = errors.New("tripId имеет неверный формат")
)

type TripResponse struct {
	ID            string `json:"id"`
	DriverID      string `json:"driverId"`
	FromPoint     string `json:"fromPoint"`
	ToPoint       string `json:"toPoint"`
	DepartureTime string `json:"departureTime"`
	Seats         int    `json:"seats"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
}

func (h *TripHandler) GetTrip(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger := logctx.Logger(ctx).With(
		slog.String("handler", "GetTrip"),
	)

	tracer := otel.Tracer("trip-api")

	ctx, span := tracer.Start(c.UserContext(), "GetTripHandler")
	defer span.End()

	c.SetUserContext(ctx)
	c.Set("trace-id", span.SpanContext().TraceID().String())
	span.SetAttributes(attribute.String("operation", "get_trip"))

	query, err := getTripQueryFromPath(c.Params("tripId"))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn(
			"получение поездки не выполнено: некорректный запрос",
			slog.Any("error", err),
		)
		return writeError(c, ErrorCodeValidation, err.Error())
	}
	span.SetAttributes(attribute.String("trip_id", query.TripID.String()))

	logger = logger.With(
		slog.String("trip_id", query.TripID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	c.SetUserContext(ctx)

	logger.Info("запрос на получение поездки принят")

	trip, err := h.tripService.GetTrip(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Error(
			"получение поездки не выполнено",
			slog.Any("error", err),
		)
		return writeServiceError(c, err)
	}

	logger.Info(
		"получение поездки завершено",
		slog.String("driver_id", trip.DriverID.String()),
		slog.String("status", string(trip.Status)),
	)
	span.SetAttributes(
		attribute.String("trip_id", trip.ID.String()),
		attribute.String("driver_id", trip.DriverID.String()),
		attribute.String("status", string(trip.Status)),
	)

	return c.Status(fiber.StatusOK).JSON(tripResponseFromView(trip))
}

func getTripQueryFromPath(tripIDParam string) (service.GetTripQuery, error) {
	tripID, err := uuid.Parse(tripIDParam)
	if err != nil {
		return service.GetTripQuery{}, errInvalidTripID
	}

	return service.GetTripQuery{TripID: tripID}, nil
}

func tripResponseFromView(trip service.TripView) TripResponse {
	return TripResponse{
		ID:            trip.ID.String(),
		DriverID:      trip.DriverID.String(),
		FromPoint:     trip.FromPoint,
		ToPoint:       trip.ToPoint,
		DepartureTime: formatHTTPTime(trip.DepartureTime),
		Seats:         trip.Seats,
		Status:        string(trip.Status),
		CreatedAt:     formatHTTPTime(trip.CreatedAt),
	}
}
