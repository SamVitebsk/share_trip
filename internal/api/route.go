package api

import (
	"share_trip/internal/api/middleware"

	"github.com/gofiber/fiber/v2"
)

const clientRole = "client"

func (s *Server) Route(route fiber.Router, authMiddleware fiber.Handler, keycloakClientID string) {
	route.Get("/ready", s.readyHandler.Ready)
	route.Post(
		"/trip/create",
		authMiddleware,
		middleware.RequireClientRole(keycloakClientID, clientRole),
		s.tripHandler.CreateTrip,
	)
	route.Get(
		"/trip/:tripId",
		authMiddleware,
		middleware.RequireClientRole(keycloakClientID, clientRole),
		s.tripHandler.GetTrip,
	)
	route.Post(
		"/trip/:tripId/publish",
		authMiddleware,
		middleware.RequireClientRole(keycloakClientID, clientRole),
		s.tripHandler.PublishTrip,
	)
}
