package api

import (
	"errors"
	"strings"
	"time"

	"share_trip/internal/service"

	"github.com/google/uuid"
)

var (
	errInvalidDriverID      = errors.New("driverID имеет неверный формат")
	errInvalidDepartureTime = errors.New("departureTime имеет неверный формат")
	errInvalidTripID        = errors.New("tripId имеет неверный формат")
)

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

func parseUUIDIfPresent(value string, invalidErr error) (uuid.UUID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return uuid.Nil, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, invalidErr
	}

	return parsed, nil
}

func parseRFC3339TimeIfPresent(value string, invalidErr error) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, invalidErr
	}

	return parsed, nil
}

func formatHTTPTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format(time.RFC3339Nano)
}
