package api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

const readinessTimeout = 2 * time.Second

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

type ReadyHandler struct {
	checker ReadinessChecker
}

func NewReadyHandler(checker ReadinessChecker) *ReadyHandler {
	return &ReadyHandler{checker: checker}
}

func (h *ReadyHandler) Ready(c *fiber.Ctx) error {
	if h.checker == nil {
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}

	ctx, cancel := context.WithTimeout(c.Context(), readinessTimeout)
	defer cancel()

	if err := h.checker.Ping(ctx); err != nil {
		return c.SendStatus(fiber.StatusServiceUnavailable)
	}
	return c.SendStatus(fiber.StatusOK)
}
