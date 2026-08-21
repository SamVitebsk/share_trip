package api

import (
	"errors"
	"share_trip/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
