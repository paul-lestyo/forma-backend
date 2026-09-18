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

func (s *recapService) GetRecap(ctx context.Context, userID int64, monthStr string, weekOffset int) (*domain.RecapResponse, error) {
	totalCustom, _ := s.todoRepo.CountCompletedByUserID(ctx, userID)
	totalLogs, _ := s.logRepo.CountCompletedByUserID(ctx, userID)
	totalCompleted := totalCustom + totalLogs

	user, _ := s.userRepo.FindByID(ctx, userID)
	streakDays := CalculateUserStreak(ctx, userID, s.logRepo, s.todoRepo)
	if user != nil && user.StreakDays != streakDays {
		user.StreakDays = streakDays
		todayWIB := GetWIBTodayString()
		_ = s.userRepo.UpdateStats(ctx, user.ID, user.Level, user.CurrentEXP, user.TotalEXP, streakDays, user.TitleRank, todayWIB)
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

	// 1. 7-Day EXP History (with weekOffset support)
	if weekOffset > 0 {
		weekOffset = 0 // prevent navigating to future weeks
	}

	endDay := now.AddDate(0, 0, weekOffset*7)
	startDay := endDay.AddDate(0, 0, -6)
	weekRangeStr := fmt.Sprintf("%s - %s", startDay.Format("02 Jan"), endDay.Format("02 Jan"))

	for i := 6; i >= 0; i-- {
		day := endDay.AddDate(0, 0, -i)
		dateStr := day.Format("2006-01-02")
		dailyKey, _, _, _ := GetPeriodKeys(dateStr)

		dayEXP, _ := s.todoRepo.SumEXPByUserIDAndDate(ctx, userID, dateStr)

		for _, log := range allHabitLogs {
			if log.Completed != 1 {
				continue
			}

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

	// 2. Weekly completion rate
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

	// 3. Pre-index completed todos & habit logs by date for fast lookup
	completedTodos, _ := s.todoRepo.FindAllCompletedByUserID(ctx, userID)
	todoEXPByDate := make(map[string]int)
	todoCountByDate := make(map[string]int)
	for _, td := range completedTodos {
		todoEXPByDate[td.Date] += td.EXPReward
		todoCountByDate[td.Date]++
	}

	habitEXPByDate := make(map[string]int)
	habitCountByDate := make(map[string]int)
	for _, log := range allHabitLogs {
		if log.Completed != 1 {
			continue
		}
		tmpl, exists := tmplMap[log.TemplateID]
		if !exists {
			continue
		}

		logDate := ""
		if log.CompletedAt != "" && len(log.CompletedAt) >= 10 {
			logDate = log.CompletedAt[:10]
		} else if tmpl.Frequency == "daily" && len(log.PeriodKey) == 10 {
			logDate = log.PeriodKey
		}

		if logDate != "" {
			habitEXPByDate[logDate] += tmpl.EXPReward
			habitCountByDate[logDate]++
		}
	}

	// 4. Monthly Activity
	if monthStr == "" || len(monthStr) != 7 {
		monthStr = now.Format("2006-01")
	}

	targetMonth, err := time.ParseInLocation("2006-01", monthStr, wibLocation)
	if err != nil {
		targetMonth = now
		monthStr = now.Format("2006-01")
	}

	firstDay := time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, wibLocation)
	lastDay := firstDay.AddDate(0, 1, -1)
	daysInMonth := lastDay.Day()
	monthName := targetMonth.Format("January 2006")

	var monthlyActivity []domain.MonthlyDayActivity
	totalEXPMonth := 0

	for d := 1; d <= daysInMonth; d++ {
		currentDate := time.Date(targetMonth.Year(), targetMonth.Month(), d, 0, 0, 0, 0, wibLocation)
		dateStr := currentDate.Format("2006-01-02")

		dayEXP := todoEXPByDate[dateStr] + habitEXPByDate[dateStr]
		completedCount := todoCountByDate[dateStr] + habitCountByDate[dateStr]
		totalEXPMonth += dayEXP

		level := 0
		if dayEXP >= 100 {
			level = 4
		} else if dayEXP >= 50 {
			level = 3
		} else if dayEXP >= 25 {
			level = 2
		} else if dayEXP > 0 {
			level = 1
		}

		isoWeekday := int(currentDate.Weekday())
		if isoWeekday == 0 {
			isoWeekday = 7
		}

		monthlyActivity = append(monthlyActivity, domain.MonthlyDayActivity{
			Date:           dateStr,
			Day:            d,
			DayOfWeek:      isoWeekday,
			EXPEarned:      dayEXP,
			Level:          level,
			CompletedCount: completedCount,
		})
	}

	// 5. GitHub Contribution Grid (Past 20 weeks, Mon to Sun = 140 days)
	daysFromMon := (int(now.Weekday()) + 6) % 7
	thisWeekMon := time.Date(now.Year(), now.Month(), now.Day()-daysFromMon, 0, 0, 0, 0, wibLocation)
	thisWeekSun := thisWeekMon.AddDate(0, 0, 6)
	startMon := thisWeekMon.AddDate(0, 0, -19*7) // 20 weeks total

	var contributionGrid []domain.ContributionDay
	totalContributions := 0
	todayStr := now.Format("2006-01-02")

	for curr := startMon; !curr.After(thisWeekSun); curr = curr.AddDate(0, 0, 1) {
		dateStr := curr.Format("2006-01-02")
		isFuture := dateStr > todayStr

		dayEXP := 0
		completedCount := 0
		level := 0

		if isFuture {
			level = -1
		} else {
			dayEXP = todoEXPByDate[dateStr] + habitEXPByDate[dateStr]
			completedCount = todoCountByDate[dateStr] + habitCountByDate[dateStr]
			totalContributions += completedCount

			if dayEXP >= 100 {
				level = 4
			} else if dayEXP >= 50 {
				level = 3
			} else if dayEXP >= 25 {
				level = 2
			} else if dayEXP > 0 {
				level = 1
			}
		}

		isoWeekday := int(curr.Weekday())
		if isoWeekday == 0 {
			isoWeekday = 7
		}

		contributionGrid = append(contributionGrid, domain.ContributionDay{
			Date:           dateStr,
			Day:            curr.Day(),
			DayOfWeek:      isoWeekday,
			Month:          curr.Format("Jan"),
			EXPEarned:      dayEXP,
			Level:          level,
			CompletedCount: completedCount,
			IsFuture:       isFuture,
		})
	}

	return &domain.RecapResponse{
		TotalQuestsCompleted: totalCompleted,
		StreakDays:           streakDays,
		WeeklyCompletionRate: completionRate,
		TotalEXPThisWeek:     totalEXPThisWeek,
		PeakDay:              peakDayName,
		ActiveRoutinesCount:  activeRoutinesCount,
		RecentEXPHistory:     expHistory,
		WeekRange:            weekRangeStr,
		WeekOffset:           weekOffset,
		SelectedMonth:        monthStr,
		MonthName:            monthName,
		TotalEXPMonth:        totalEXPMonth,
		MonthlyActivity:      monthlyActivity,
		ContributionGrid:     contributionGrid,
		TotalContributions:   totalContributions,
	}, nil
}
