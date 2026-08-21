package api

import (
	"errors"
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
	var req CreateTripRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, ErrorCodeValidation, "тело запроса должно быть валидным JSON")
	}

	cmd, err := createTripCommandFromRequest(req)
	if err != nil {
		return writeError(c, ErrorCodeValidation, err.Error())
	}

	result, err := h.tripService.CreateTrip(c.Context(), cmd)
	if err != nil {
		return writeServiceError(c, err)
	}

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
