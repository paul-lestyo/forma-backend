package fiberadapter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	fibermiddleware "habittracker-be/internal/adapters/http/fiber/middleware"
)

func SetupRouter(
	authHandler *AuthHandler,
	trackerHandler *TrackerHandler,
	routineHandler *RoutineHandler,
	recapHandler *RecapHandler,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Habit Tracker API v1.0 (Hexagonal)",
		ServerHeader: "Go-Fiber",
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: false,
	}))

	// Health Check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "Habit Tracker Backend",
			"engine":  "Go Fiber v2 + SQLite (Hexagonal Architecture)",
		})
	})

	// Public Auth Routes
	api := app.Group("/api")
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected Routes (JWT required)
	protected := api.Group("", fibermiddleware.Protected())

	// Auth Profile
	protected.Get("/auth/me", authHandler.GetMe)

	// Tracker Routes
	protected.Get("/tracker", trackerHandler.GetTracker)
	protected.Post("/todos", trackerHandler.CreateCustomTodo)
	protected.Patch("/todos/toggle", trackerHandler.ToggleQuest)
	protected.Delete("/todos/:id", trackerHandler.DeleteQuest)

	// Habit Templates (Routine Tab)
	protected.Get("/templates", routineHandler.GetTemplates)
	protected.Post("/templates", routineHandler.CreateTemplate)
	protected.Put("/templates/:id", routineHandler.UpdateTemplate)
	protected.Delete("/templates/:id", routineHandler.DeleteTemplate)

	// Recap Analytics
	protected.Get("/recap", recapHandler.GetRecap)

	return app
}
