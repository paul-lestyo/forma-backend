package ports

import (
	"context"
	"habittracker-be/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateStats(ctx context.Context, userID int64, level, currentEXP, totalEXP int, titleRank, lastActiveDate string) error
}

type HabitRepository interface {
	Create(ctx context.Context, tmpl *domain.HabitTemplate) error
	FindActiveByUserID(ctx context.Context, userID int64) ([]domain.HabitTemplate, error)
	FindAllByUserID(ctx context.Context, userID int64) ([]domain.HabitTemplate, error)
	FindByID(ctx context.Context, id int64, userID int64) (*domain.HabitTemplate, error)
	Update(ctx context.Context, tmpl *domain.HabitTemplate) error
	Delete(ctx context.Context, id int64, userID int64) error
	SoftDelete(ctx context.Context, id int64, userID int64) error
}

type TodoRepository interface {
	Create(ctx context.Context, todo *domain.CustomTodo) error
	FindByUserIDAndDate(ctx context.Context, userID int64, date string) ([]domain.CustomTodo, error)
	FindByID(ctx context.Context, id int64, userID int64) (*domain.CustomTodo, error)
	UpdateCompletion(ctx context.Context, id int64, completed int) error
	Delete(ctx context.Context, id int64, userID int64) error
	CountCompletedByUserID(ctx context.Context, userID int64) (int, error)
	SumEXPByUserIDAndDate(ctx context.Context, userID int64, date string) (int, error)
}

type HabitLogRepository interface {
	Create(ctx context.Context, log *domain.HabitLog) error
	FindByUserTemplatePeriod(ctx context.Context, userID, templateID int64, periodKey string) (*domain.HabitLog, error)
	DeleteByUserTemplatePeriod(ctx context.Context, userID, templateID int64, periodKey string) error
	CountCompletedByUserID(ctx context.Context, userID int64) (int, error)
	FindAllByUserID(ctx context.Context, userID int64) ([]domain.HabitLog, error)
}
