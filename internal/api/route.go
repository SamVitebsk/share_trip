package api

import "github.com/gofiber/fiber/v2"

func (s *Server) Route(route fiber.Router) {
	route.Get("/ready", s.readyHandler.Ready)
	route.Post("/trip/create", s.tripHandler.CreateTrip)
	route.Get("/trip/:tripId", s.tripHandler.GetTrip)
}
