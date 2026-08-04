package api

import (
	"github.com/gofiber/fiber/v2"
)

func (s *Server) Ready(c *fiber.Ctx) error {
	err := s.Repository.Ping(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}
