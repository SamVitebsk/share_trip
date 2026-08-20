package api

import (
	"share_trip/internal/service"

	"github.com/gofiber/fiber/v2"
)

type TripHandler struct {
	tripService *service.TripService
}

func NewTripHandler(tripService *service.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
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

func (h *TripHandler) GetTrip(c *fiber.Ctx) error {
	query, err := getTripQueryFromPath(c.Params("tripId"))
	if err != nil {
		return writeError(c, ErrorCodeValidation, err.Error())
	}

	trip, err := h.tripService.GetTrip(c.Context(), query)
	if err != nil {
		return writeServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(tripResponseFromView(trip))
}
