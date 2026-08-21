package fiberadapter

import (
	"github.com/gofiber/fiber/v2"

	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

type TrackerHandler struct {
	trackerService ports.TrackerService
}

func NewTrackerHandler(trackerService ports.TrackerService) *TrackerHandler {
	return &TrackerHandler{trackerService: trackerService}
}

func getUserID(c *fiber.Ctx) (int64, bool) {
	val := c.Locals("userID")
	if val == nil {
		return 0, false
	}
	userID, ok := val.(int64)
	if !ok || userID <= 0 {
		return 0, false
	}
	return userID, true
}

func (h *TrackerHandler) GetTracker(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}

	dateParam := c.Query("date")

	quests, err := h.trackerService.GetTrackerQuests(c.Context(), userID, dateParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if quests == nil {
		quests = []domain.QuestItem{}
	}

	return c.JSON(fiber.Map{
		"date":   dateParam,
		"quests": quests,
	})
}

func (h *TrackerHandler) CreateCustomTodo(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}

	var req domain.CreateCustomTodoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	quest, err := h.trackerService.CreateCustomQuest(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(quest)
}

func (h *TrackerHandler) ToggleQuest(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}

	var req domain.ToggleQuestRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	res, err := h.trackerService.ToggleQuest(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(res)
}

func (h *TrackerHandler) DeleteQuest(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
	}

	questID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid quest ID"})
	}

	itemType := c.Query("type", "custom")

	if err := h.trackerService.DeleteQuest(c.Context(), userID, int64(questID), itemType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Quest deleted successfully"})
}
