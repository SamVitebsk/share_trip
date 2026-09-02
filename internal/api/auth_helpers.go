package api

import (
	"strings"

	"share_trip/internal/api/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func driverIDFromClaims(c *fiber.Ctx) (uuid.UUID, error) {
	claims, err := middleware.ClaimsFromContext(c)
	if err != nil {
		return uuid.Nil, err
	}

	driverID, err := uuid.Parse(strings.TrimSpace(claims.Subject))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "subject токена имеет неверный формат")
	}

	return driverID, nil
}
