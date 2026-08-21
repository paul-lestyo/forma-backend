package fiberadapter

import (
	"github.com/gofiber/fiber/v2"

	"habittracker-be/internal/core/ports"
)

type RecapHandler struct {
	recapService ports.RecapService
}

func NewRecapHandler(recapService ports.RecapService) *RecapHandler {
	return &RecapHandler{recapService: recapService}
}

func (h *RecapHandler) GetRecap(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}

	res, err := h.recapService.GetRecap(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}
