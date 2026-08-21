package api

import (
	"share_trip/internal/service"
)

type TripHandler struct {
	tripService *service.TripService
}

func NewTripHandler(tripService *service.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}
