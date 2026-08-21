package sqlite

import (
	"context"
	"database/sql"
	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

type logRepo struct {
	db *sql.DB
}

func NewHabitLogRepository(db *sql.DB) ports.HabitLogRepository {
	return &logRepo{db: db}
}

func (r *logRepo) Create(ctx context.Context, log *domain.HabitLog) error {
	res, err := r.db.ExecContext(
		ctx,
		"INSERT INTO habit_logs (user_id, template_id, period_key, completed) VALUES (?, ?, ?, 1)",
		log.UserID, log.TemplateID, log.PeriodKey,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		log.ID = id
		log.Completed = 1
	}
	return nil
}

func (r *logRepo) FindByUserTemplatePeriod(ctx context.Context, userID, templateID int64, periodKey string) (*domain.HabitLog, error) {
	var l domain.HabitLog
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, template_id, period_key, completed, completed_at FROM habit_logs WHERE user_id = ? AND template_id = ? AND period_key = ?",
		userID, templateID, periodKey,
	).Scan(&l.ID, &l.UserID, &l.TemplateID, &l.PeriodKey, &l.Completed, &l.CompletedAt)

	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *logRepo) DeleteByUserTemplatePeriod(ctx context.Context, userID, templateID int64, periodKey string) error {
	_, err := r.db.ExecContext(
		ctx,
		"DELETE FROM habit_logs WHERE user_id = ? AND template_id = ? AND period_key = ?",
		userID, templateID, periodKey,
	)
	return err
}

func (r *logRepo) CountCompletedByUserID(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM habit_logs WHERE user_id = ? AND completed = 1", userID).Scan(&count)
	return count, err
}

func (r *logRepo) FindAllByUserID(ctx context.Context, userID int64) ([]domain.HabitLog, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, user_id, template_id, period_key, completed, completed_at FROM habit_logs WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.HabitLog
	for rows.Next() {
		var l domain.HabitLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.TemplateID, &l.PeriodKey, &l.Completed, &l.CompletedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}
	return logs, nil
}
