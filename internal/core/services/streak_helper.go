package services

import (
	"context"
	"strings"
	"time"

	"habittracker-be/internal/core/ports"
)

// CalculateUserStreak calculates consecutive days with at least 1 completed task (habit or custom todo).
// If today has activity, streak includes today.
// If today has no activity yet, it checks yesterday: if yesterday was active, the streak is kept alive.
// If neither today nor yesterday had activity, streak is 0.
func CalculateUserStreak(
	ctx context.Context,
	userID int64,
	logRepo ports.HabitLogRepository,
	todoRepo ports.TodoRepository,
) int {
	todayStr := GetWIBTodayString()
	todayTime, err := time.ParseInLocation("2006-01-02", todayStr, wibLocation)
	if err != nil {
		return 0
	}

	completedDates := make(map[string]bool)

	// 1. Collect completed dates from habit_logs
	allLogs, err := logRepo.FindAllByUserID(ctx, userID)
	if err == nil {
		for _, log := range allLogs {
			if log.Completed != 1 {
				continue
			}
			if log.CompletedAt != "" && len(log.CompletedAt) >= 10 {
				completedDates[log.CompletedAt[:10]] = true
			}
			// If daily habit, period_key is YYYY-MM-DD
			if len(log.PeriodKey) == 10 && strings.Count(log.PeriodKey, "-") == 2 {
				completedDates[log.PeriodKey] = true
			}
		}
	}

	// 2. Collect completed dates from custom_todos
	if todoRepo != nil {
		todoDates, err := todoRepo.FindCompletedDatesByUserID(ctx, userID)
		if err == nil {
			for _, d := range todoDates {
				completedDates[d] = true
			}
		}
	}

	streak := 0

	// Case A: Today already has at least 1 completed activity
	if completedDates[todayStr] {
		checkDate := todayTime
		for {
			dateStr := checkDate.Format("2006-01-02")
			if completedDates[dateStr] {
				streak++
				checkDate = checkDate.AddDate(0, 0, -1)
			} else {
				break
			}
		}
		return streak
	}

	// Case B: Today has no activity yet, check yesterday to keep streak alive
	yesterdayTime := todayTime.AddDate(0, 0, -1)
	yesterdayStr := yesterdayTime.Format("2006-01-02")
	if completedDates[yesterdayStr] {
		checkDate := yesterdayTime
		for {
			dateStr := checkDate.Format("2006-01-02")
			if completedDates[dateStr] {
				streak++
				checkDate = checkDate.AddDate(0, 0, -1)
			} else {
				break
			}
		}
		return streak
	}

	// Case C: Neither today nor yesterday had activity -> streak broken
	return 0
}
