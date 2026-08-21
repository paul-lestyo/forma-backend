package ports

import (
	"context"
	"habittracker-be/internal/core/domain"
)

type AuthService interface {
	Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error)
	GetProfile(ctx context.Context, userID int64) (*domain.User, error)
}

type TrackerService interface {
	GetTrackerQuests(ctx context.Context, userID int64, dateStr string) ([]domain.QuestItem, error)
	CreateCustomQuest(ctx context.Context, userID int64, req domain.CreateCustomTodoRequest) (*domain.QuestItem, error)
	ToggleQuest(ctx context.Context, userID int64, req domain.ToggleQuestRequest) (*domain.ToggleQuestResponse, error)
	DeleteQuest(ctx context.Context, userID int64, id int64, itemType string) error
}

type RoutineService interface {
	GetTemplates(ctx context.Context, userID int64) ([]domain.HabitTemplate, error)
	CreateTemplate(ctx context.Context, userID int64, req domain.CreateTemplateRequest) (*domain.HabitTemplate, error)
	UpdateTemplate(ctx context.Context, userID int64, id int64, req domain.CreateTemplateRequest) error
	DeleteTemplate(ctx context.Context, userID int64, id int64) error
}

type RecapService interface {
	GetRecap(ctx context.Context, userID int64) (*domain.RecapResponse, error)
}
