package sqlite

import (
	"context"
	"database/sql"
	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

type habitRepo struct {
	db *sql.DB
}

func NewHabitRepository(db *sql.DB) ports.HabitRepository {
	return &habitRepo{db: db}
}

func (r *habitRepo) Create(ctx context.Context, tmpl *domain.HabitTemplate) error {
	res, err := r.db.ExecContext(
		ctx,
		"INSERT INTO habit_templates (user_id, title, priority, exp_reward, frequency, is_active) VALUES (?, ?, ?, ?, ?, 1)",
		tmpl.UserID, tmpl.Title, tmpl.Priority, tmpl.EXPReward, tmpl.Frequency,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		tmpl.ID = id
		tmpl.IsActive = 1
	}
	return nil
}

func (r *habitRepo) FindActiveByUserID(ctx context.Context, userID int64) ([]domain.HabitTemplate, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, user_id, title, priority, exp_reward, frequency, is_active, created_at FROM habit_templates WHERE user_id = ? AND is_active = 1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.HabitTemplate
	for rows.Next() {
		var tmpl domain.HabitTemplate
		if err := rows.Scan(&tmpl.ID, &tmpl.UserID, &tmpl.Title, &tmpl.Priority, &tmpl.EXPReward, &tmpl.Frequency, &tmpl.IsActive, &tmpl.CreatedAt); err != nil {
			continue
		}
		templates = append(templates, tmpl)
	}
	return templates, nil
}

func (r *habitRepo) FindAllByUserID(ctx context.Context, userID int64) ([]domain.HabitTemplate, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, user_id, title, priority, exp_reward, frequency, is_active, created_at FROM habit_templates WHERE user_id = ? ORDER BY id DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []domain.HabitTemplate
	for rows.Next() {
		var tmpl domain.HabitTemplate
		if err := rows.Scan(&tmpl.ID, &tmpl.UserID, &tmpl.Title, &tmpl.Priority, &tmpl.EXPReward, &tmpl.Frequency, &tmpl.IsActive, &tmpl.CreatedAt); err != nil {
			continue
		}
		templates = append(templates, tmpl)
	}
	return templates, nil
}

func (r *habitRepo) FindByID(ctx context.Context, id int64, userID int64) (*domain.HabitTemplate, error) {
	var tmpl domain.HabitTemplate
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, title, priority, exp_reward, frequency, is_active, created_at FROM habit_templates WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&tmpl.ID, &tmpl.UserID, &tmpl.Title, &tmpl.Priority, &tmpl.EXPReward, &tmpl.Frequency, &tmpl.IsActive, &tmpl.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func (r *habitRepo) Update(ctx context.Context, tmpl *domain.HabitTemplate) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE habit_templates SET title = ?, priority = ?, exp_reward = ?, frequency = ? WHERE id = ? AND user_id = ?",
		tmpl.Title, tmpl.Priority, tmpl.EXPReward, tmpl.Frequency, tmpl.ID, tmpl.UserID,
	)
	return err
}

func (r *habitRepo) Delete(ctx context.Context, id int64, userID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM habit_templates WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func (r *habitRepo) SoftDelete(ctx context.Context, id int64, userID int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE habit_templates SET is_active = 0 WHERE id = ? AND user_id = ?", id, userID)
	return err
}
