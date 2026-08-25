package api

import (
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
	var req PublishTripRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, ErrorCodeValidation, "тело запроса должно быть валидным JSON")
	}

	publishTripCommand, err := parsePublishTripCommandFromRequest(c.Params("tripId"), req)
	if err != nil {
		return writeError(c, ErrorCodeValidation, err.Error())
	}

	result, err := h.tripService.PublishTrip(c.Context(), *publishTripCommand)
	if err != nil {
		return writeServiceError(c, err)
	}

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
