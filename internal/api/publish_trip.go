package api

import (
	"share_trip/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PublishTripRequest struct {
	DriverID string `json:"driverId"`
}

func (h *TripHandler) PublishTrip(c *fiber.Ctx) error {
	var req PublishTripRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, ErrorCodeValidation, "тело запроса должно быть валидным JSON")
	}

	cmd, err := publishTripCommandFromRequest(c.Params("tripId"), req)
	if err != nil {
		return writeError(c, ErrorCodeValidation, err.Error())
	}

	if err := h.tripService.PublishTrip(c.Context(), cmd); err != nil {
		return writeServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func publishTripCommandFromRequest(tripIDParam string, req PublishTripRequest) (service.PublishTripCommand, error) {
	tripID, err := uuid.Parse(tripIDParam)
	if err != nil {
		return service.PublishTripCommand{}, errInvalidTripID
	}

	driverID, err := parseUUIDIfPresent(req.DriverID, errInvalidDriverID)
	if err != nil {
		return service.PublishTripCommand{}, err
	}

	return service.PublishTripCommand{
		TripID:   tripID,
		DriverID: driverID,
	}, nil
}
