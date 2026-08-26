package api

import (
	"errors"
	"log/slog"
	"share_trip/internal/observability/logctx"
	"share_trip/internal/service"

	"github.com/gofiber/fiber/v2"
)

var (
	errInvalidDriverID      = errors.New("driverID имеет неверный формат")
	errInvalidDepartureTime = errors.New("departureTime имеет неверный формат")
)

type CreateTripRequest struct {
	DriverID      string `json:"driverId"`
	FromPoint     string `json:"fromPoint"`
	ToPoint       string `json:"toPoint"`
	DepartureTime string `json:"departureTime"`
	Seats         int    `json:"seats"`
}

type CreateTripResponse struct {
	ID string `json:"id"`
}

func (h *TripHandler) CreateTrip(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger := logctx.Logger(ctx).With(
		slog.String("handler", "CreateTrip"),
	)

	var req CreateTripRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Warn(
			"создание поездки не выполнено: некорректный JSON в теле запроса",
			slog.Any("error", err),
		)
		return writeError(c, ErrorCodeValidation, "тело запроса должно быть валидным JSON")
	}

	createTripCommand, err := createTripCommandFromRequest(req)
	if err != nil {
		logger.Warn(
			"создание поездки не выполнено: некорректный запрос",
			slog.Any("error", err),
		)
		return writeError(c, ErrorCodeValidation, err.Error())
	}

	logger = logger.With(
		slog.String("driver_id", createTripCommand.DriverID.String()),
	)
	ctx = logctx.WithLogger(ctx, logger)
	c.SetUserContext(ctx)

	logger.Info("запрос на создание поездки принят")

	result, err := h.tripService.CreateTrip(ctx, createTripCommand)
	if err != nil {
		logger.Error(
			"создание поездки не выполнено",
			slog.Any("error", err),
		)
		return writeServiceError(c, err)
	}

	logger.Info(
		"создание поездки завершено",
		slog.String("trip_id", result.TripID.String()),
	)

	return c.Status(fiber.StatusCreated).JSON(createTripResponseFromResult(result))
}

func createTripCommandFromRequest(req CreateTripRequest) (service.CreateTripCommand, error) {
	driverID, err := parseUUIDIfPresent(req.DriverID, errInvalidDriverID)
	if err != nil {
		return service.CreateTripCommand{}, err
	}

	departureTime, err := parseRFC3339TimeIfPresent(req.DepartureTime, errInvalidDepartureTime)
	if err != nil {
		return service.CreateTripCommand{}, err
	}

	return service.CreateTripCommand{
		DriverID:      driverID,
		FromPoint:     req.FromPoint,
		ToPoint:       req.ToPoint,
		DepartureTime: departureTime,
		Seats:         req.Seats,
	}, nil
}

func createTripResponseFromResult(result service.CreateTripResult) CreateTripResponse {
	return CreateTripResponse{
		ID: result.TripID.String(),
	}
}
