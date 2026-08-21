package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

type recapService struct {
	userRepo  ports.UserRepository
	habitRepo ports.HabitRepository
	todoRepo  ports.TodoRepository
	logRepo   ports.HabitLogRepository
}

func NewRecapService(
	userRepo ports.UserRepository,
	habitRepo ports.HabitRepository,
	todoRepo ports.TodoRepository,
	logRepo ports.HabitLogRepository,
) ports.RecapService {
	return &recapService{
		userRepo:  userRepo,
		habitRepo: habitRepo,
		todoRepo:  todoRepo,
		logRepo:   logRepo,
	}
}

func (s *recapService) GetRecap(ctx context.Context, userID int64) (*domain.RecapResponse, error) {
	totalCustom, _ := s.todoRepo.CountCompletedByUserID(ctx, userID)
	totalLogs, _ := s.logRepo.CountCompletedByUserID(ctx, userID)
	totalCompleted := totalCustom + totalLogs

	user, _ := s.userRepo.FindByID(ctx, userID)
	streakDays := 0
	if user != nil {
		streakDays = user.StreakDays
	}

	now := time.Now().In(wibLocation)
	var expHistory []domain.EXPDayHistory

	allHabitLogs, _ := s.logRepo.FindAllByUserID(ctx, userID)
	allTemplates, _ := s.habitRepo.FindAllByUserID(ctx, userID)
	activeRoutinesCount := len(allTemplates)

	tmplMap := make(map[int64]domain.HabitTemplate)
	for _, t := range allTemplates {
		tmplMap[t.ID] = t
	}

	totalEXPThisWeek := 0
	maxDayEXP := -1
	peakDayName := "N/A"

	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")
		dailyKey, _, _, _ := GetPeriodKeys(dateStr)

		dayEXP, _ := s.todoRepo.SumEXPByUserIDAndDate(ctx, userID, dateStr)

		for _, log := range allHabitLogs {
			tmpl, exists := tmplMap[log.TemplateID]
			if !exists {
				continue
			}

			// Attrib EXP to exact day completed
			completedOnThisDay := false
			if log.CompletedAt != "" && strings.HasPrefix(log.CompletedAt, dateStr) {
				completedOnThisDay = true
			} else if tmpl.Frequency == "daily" && log.PeriodKey == dailyKey {
				completedOnThisDay = true
			}

			if completedOnThisDay {
				dayEXP += tmpl.EXPReward
			}
		}

		totalEXPThisWeek += dayEXP
		if dayEXP > maxDayEXP && dayEXP > 0 {
			maxDayEXP = dayEXP
			peakDayName = fmt.Sprintf("%s (+%d EXP)", day.Format("Monday"), dayEXP)
		}

		expHistory = append(expHistory, domain.EXPDayHistory{
			Date:      day.Format("02 Jan"),
			EXPEarned: dayEXP,
		})
	}

	if peakDayName == "N/A" && len(expHistory) > 0 {
		peakDayName = "Today"
	}

	// Calculate weekly completion rate
	completionRate := 100.0
	if activeRoutinesCount > 0 {
		expectedWeeklyTasks := 0
		for _, t := range allTemplates {
			if t.Frequency == "daily" {
				expectedWeeklyTasks += 7
			} else {
				expectedWeeklyTasks += 1
			}
		}
		if expectedWeeklyTasks > 0 {
			completionRate = (float64(totalCompleted) / float64(expectedWeeklyTasks)) * 100.0
			if completionRate > 100.0 {
				completionRate = 100.0
			}
		}
	} else if totalCompleted == 0 {
		completionRate = 0.0
	}

	return &domain.RecapResponse{
		TotalQuestsCompleted: totalCompleted,
		StreakDays:           streakDays,
		WeeklyCompletionRate: completionRate,
		TotalEXPThisWeek:     totalEXPThisWeek,
		PeakDay:              peakDayName,
		ActiveRoutinesCount:  activeRoutinesCount,
		RecentEXPHistory:     expHistory,
	}, nil
}
