package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

var wibLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	wibLocation = loc
}

func GetWIBTodayString() string {
	return time.Now().In(wibLocation).Format("2006-01-02")
}

func GetPeriodKeys(dateStr string) (dailyKey string, weeklyKey string, monthlyKey string, err error) {
	t, err := time.ParseInLocation("2006-01-02", dateStr, wibLocation)
	if err != nil {
		return "", "", "", err
	}
	dailyKey = dateStr
	year, week := t.ISOWeek()
	weeklyKey = fmt.Sprintf("%04d-W%02d", year, week)
	monthlyKey = t.Format("2006-01")
	return dailyKey, weeklyKey, monthlyKey, nil
}

type trackerService struct {
	userRepo  ports.UserRepository
	habitRepo ports.HabitRepository
	todoRepo  ports.TodoRepository
	logRepo   ports.HabitLogRepository
}

func NewTrackerService(
	userRepo ports.UserRepository,
	habitRepo ports.HabitRepository,
	todoRepo ports.TodoRepository,
	logRepo ports.HabitLogRepository,
) ports.TrackerService {
	return &trackerService{
		userRepo:  userRepo,
		habitRepo: habitRepo,
		todoRepo:  todoRepo,
		logRepo:   logRepo,
	}
}

func (s *trackerService) GetTrackerQuests(ctx context.Context, userID int64, dateStr string) ([]domain.QuestItem, error) {
	if dateStr == "" {
		dateStr = GetWIBTodayString()
	}

	dailyKey, weeklyKey, monthlyKey, err := GetPeriodKeys(dateStr)
	if err != nil {
		return nil, errors.New("invalid date format, must be YYYY-MM-DD")
	}

	var quests []domain.QuestItem

	// 1. Fetch Habit Templates
	templates, err := s.habitRepo.FindActiveByUserID(ctx, userID)
	if err == nil {
		for _, tmpl := range templates {
			targetPeriodKey := dailyKey
			if tmpl.Frequency == "weekly" {
				targetPeriodKey = weeklyKey
			} else if tmpl.Frequency == "monthly" {
				targetPeriodKey = monthlyKey
			}

			log, logErr := s.logRepo.FindByUserTemplatePeriod(ctx, userID, tmpl.ID, targetPeriodKey)
			isCompleted := (logErr == nil && log != nil && log.Completed == 1)

			quests = append(quests, domain.QuestItem{
				ID:         tmpl.ID,
				TemplateID: tmpl.ID,
				ItemType:   "habit",
				Title:      tmpl.Title,
				Priority:   tmpl.Priority,
				EXPReward:  tmpl.EXPReward,
				Frequency:  tmpl.Frequency,
				Date:       dateStr,
				Completed:  isCompleted,
			})
		}
	}

	// 2. Fetch Custom Todos for specific date
	todos, err := s.todoRepo.FindByUserIDAndDate(ctx, userID, dateStr)
	if err == nil {
		for _, todo := range todos {
			quests = append(quests, domain.QuestItem{
				ID:        todo.ID,
				ItemType:  "custom",
				Title:     todo.Title,
				Priority:  todo.Priority,
				EXPReward: todo.EXPReward,
				Date:      todo.Date,
				Completed: todo.Completed == 1,
			})
		}
	}

	return quests, nil
}

func (s *trackerService) CreateCustomQuest(ctx context.Context, userID int64, req domain.CreateCustomTodoRequest) (*domain.QuestItem, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if req.Date == "" {
		req.Date = GetWIBTodayString()
	}
	if req.Priority == "" {
		req.Priority = "MEDIUM"
	}
	if req.EXPReward <= 0 {
		switch req.Priority {
		case "LOW":
			req.EXPReward = 10
		case "HIGH":
			req.EXPReward = 50
		default:
			req.EXPReward = 25
		}
	}

	todo := &domain.CustomTodo{
		UserID:    userID,
		Title:     req.Title,
		Priority:  req.Priority,
		EXPReward: req.EXPReward,
		Date:      req.Date,
		Completed: 0,
	}

	if err := s.todoRepo.Create(ctx, todo); err != nil {
		return nil, errors.New("failed to create custom quest")
	}

	return &domain.QuestItem{
		ID:        todo.ID,
		ItemType:  "custom",
		Title:     todo.Title,
		Priority:  todo.Priority,
		EXPReward: todo.EXPReward,
		Date:      todo.Date,
		Completed: false,
	}, nil
}

func (s *trackerService) ToggleQuest(ctx context.Context, userID int64, req domain.ToggleQuestRequest) (*domain.ToggleQuestResponse, error) {
	if req.Date == "" {
		req.Date = GetWIBTodayString()
	}

	dailyKey, weeklyKey, monthlyKey, err := GetPeriodKeys(req.Date)
	if err != nil {
		return nil, errors.New("invalid date format")
	}

	var expChange int
	var isNowCompleted bool

	if req.ItemType == "custom" {
		todo, err := s.todoRepo.FindByID(ctx, req.ID, userID)
		if err != nil || todo == nil {
			return nil, errors.New("custom quest not found")
		}

		if todo.Completed == 1 {
			_ = s.todoRepo.UpdateCompletion(ctx, req.ID, 0)
			expChange = -todo.EXPReward
			isNowCompleted = false
		} else {
			_ = s.todoRepo.UpdateCompletion(ctx, req.ID, 1)
			expChange = todo.EXPReward
			isNowCompleted = true
		}

	} else { // Habit template
		templateID := req.ID
		if req.TemplateID > 0 {
			templateID = req.TemplateID
		}

		tmpl, err := s.habitRepo.FindByID(ctx, templateID, userID)
		if err != nil || tmpl == nil {
			return nil, errors.New("habit template not found")
		}

		targetPeriodKey := dailyKey
		if tmpl.Frequency == "weekly" {
			targetPeriodKey = weeklyKey
		} else if tmpl.Frequency == "monthly" {
			targetPeriodKey = monthlyKey
		}

		existingLog, _ := s.logRepo.FindByUserTemplatePeriod(ctx, userID, templateID, targetPeriodKey)

		if existingLog != nil {
			_ = s.logRepo.DeleteByUserTemplatePeriod(ctx, userID, templateID, targetPeriodKey)
			expChange = -tmpl.EXPReward
			isNowCompleted = false
		} else {
			_ = s.logRepo.Create(ctx, &domain.HabitLog{
				UserID:     userID,
				TemplateID: templateID,
				PeriodKey:  targetPeriodKey,
				Completed:  1,
			})
			expChange = tmpl.EXPReward
			isNowCompleted = true
		}
	}

	// Update user EXP and Level progression
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.CurrentEXP += expChange
	user.TotalEXP += expChange
	if user.CurrentEXP < 0 {
		user.CurrentEXP = 0
	}
	if user.TotalEXP < 0 {
		user.TotalEXP = 0
	}

	levelUp := false
	targetEXP := user.Level * 100

	for user.CurrentEXP >= targetEXP {
		user.CurrentEXP -= targetEXP
		user.Level++
		levelUp = true
		targetEXP = user.Level * 100
	}

	// Rank title updates
	if user.Level >= 15 {
		user.TitleRank = "Legendary Hero"
	} else if user.Level >= 10 {
		user.TitleRank = "Grandmaster"
	} else if user.Level >= 6 {
		user.TitleRank = "Master of Discipline"
	} else if user.Level >= 3 {
		user.TitleRank = "Apprentice Tracker"
	}

	todayWIB := GetWIBTodayString()
	_ = s.userRepo.UpdateStats(ctx, userID, user.Level, user.CurrentEXP, user.TotalEXP, user.TitleRank, todayWIB)

	return &domain.ToggleQuestResponse{
		Completed:  isNowCompleted,
		EXPGained:  expChange,
		LevelUp:    levelUp,
		NewLevel:   user.Level,
		CurrentEXP: user.CurrentEXP,
		TotalEXP:   user.TotalEXP,
		TargetEXP:  targetEXP,
	}, nil
}

func (s *trackerService) DeleteQuest(ctx context.Context, userID int64, id int64, itemType string) error {
	if itemType == "custom" {
		return s.todoRepo.Delete(ctx, id, userID)
	}
	return s.habitRepo.SoftDelete(ctx, id, userID)
}
