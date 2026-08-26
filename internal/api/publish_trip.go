package api

import (
	"log/slog"
	"share_trip/internal/observability/logctx"
	"share_trip/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PublishTripRequest struct {
	DriverID string `json:"driverId"`
}

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

	var req PublishTripRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warn(
			"публикация поездки не выполнена: некорректный JSON в теле запроса",
			slog.Any("error", err),
		)
		return writeError(c, ErrorCodeValidation, "тело запроса должно быть валидным JSON")
	}

	publishTripCommand, err := parsePublishTripCommandFromRequest(c.Params("tripId"), req)
	if err != nil {
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
	ctx = logctx.WithLogger(ctx, logger)
	c.SetUserContext(ctx)

	logger.Info("запрос на публикацию поездки принят")

	result, err := h.tripService.PublishTrip(ctx, *publishTripCommand)
	if err != nil {
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

	return c.Status(fiber.StatusOK).JSON(toPublishTripResponse(result))
}

func parsePublishTripCommandFromRequest(tripIDParam string, req PublishTripRequest) (*service.PublishTripCommand, error) {
	tripID, err := uuid.Parse(tripIDParam)
	if err != nil {
		return nil, errInvalidTripID
	}

	driverID, err := parseUUIDIfPresent(req.DriverID, errInvalidDriverID)
	if err != nil {
		return nil, err
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
