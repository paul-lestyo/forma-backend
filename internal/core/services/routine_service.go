package services

import (
	"context"
	"errors"

	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

type routineService struct {
	habitRepo ports.HabitRepository
}

func NewRoutineService(habitRepo ports.HabitRepository) ports.RoutineService {
	return &routineService{habitRepo: habitRepo}
}

func (s *routineService) GetTemplates(ctx context.Context, userID int64) ([]domain.HabitTemplate, error) {
	return s.habitRepo.FindAllByUserID(ctx, userID)
}

func (s *routineService) CreateTemplate(ctx context.Context, userID int64, req domain.CreateTemplateRequest) (*domain.HabitTemplate, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	if req.Frequency != "daily" && req.Frequency != "weekly" && req.Frequency != "monthly" {
		req.Frequency = "daily"
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

	tmpl := &domain.HabitTemplate{
		UserID:    userID,
		Title:     req.Title,
		Priority:  req.Priority,
		EXPReward: req.EXPReward,
		Frequency: req.Frequency,
		IsActive:  1,
	}

	if err := s.habitRepo.Create(ctx, tmpl); err != nil {
		return nil, errors.New("failed to create habit template")
	}

	return tmpl, nil
}

func (s *routineService) UpdateTemplate(ctx context.Context, userID int64, id int64, req domain.CreateTemplateRequest) error {
	tmpl, err := s.habitRepo.FindByID(ctx, id, userID)
	if err != nil || tmpl == nil {
		return errors.New("template not found")
	}

	tmpl.Title = req.Title
	tmpl.Priority = req.Priority
	tmpl.EXPReward = req.EXPReward
	tmpl.Frequency = req.Frequency

	return s.habitRepo.Update(ctx, tmpl)
}

func (s *routineService) DeleteTemplate(ctx context.Context, userID int64, id int64) error {
	return s.habitRepo.Delete(ctx, id, userID)
}
