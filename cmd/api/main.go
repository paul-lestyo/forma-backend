package main

import (
	"log"

	sqliterepo "habittracker-be/internal/adapters/repository/sqlite"
	fiberadapter "habittracker-be/internal/adapters/http/fiber"
	"habittracker-be/internal/core/services"
)

func main() {
	// 1. Initialize SQLite Database Adapter
	db, err := sqliterepo.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 2. Instantiate Repository Adapters
	userRepo := sqliterepo.NewUserRepository(db)
	habitRepo := sqliterepo.NewHabitRepository(db)
	todoRepo := sqliterepo.NewTodoRepository(db)
	logRepo := sqliterepo.NewHabitLogRepository(db)

	// 3. Instantiate Core Domain Services
	authService := services.NewAuthService(userRepo, habitRepo)
	trackerService := services.NewTrackerService(userRepo, habitRepo, todoRepo, logRepo)
	routineService := services.NewRoutineService(habitRepo)
	recapService := services.NewRecapService(userRepo, habitRepo, todoRepo, logRepo)

	// 4. Instantiate Fiber Handler Adapters
	authHandler := fiberadapter.NewAuthHandler(authService)
	trackerHandler := fiberadapter.NewTrackerHandler(trackerService)
	routineHandler := fiberadapter.NewRoutineHandler(routineService)
	recapHandler := fiberadapter.NewRecapHandler(recapService)

	// 5. Setup Router & Start Server
	app := fiberadapter.SetupRouter(authHandler, trackerHandler, routineHandler, recapHandler)

	log.Println("🚀 Habit Tracker Backend (Hexagonal Architecture) Server starting on port 8088...")
	if err := app.Listen(":8088"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
