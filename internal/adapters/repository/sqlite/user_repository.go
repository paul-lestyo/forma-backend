package sqlite

import (
	"context"
	"database/sql"
	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) ports.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	res, err := r.db.ExecContext(
		ctx,
		"INSERT INTO users (username, password_hash, display_name, title_rank, level, current_exp, total_exp) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user.Username, user.PasswordHash, user.DisplayName, user.TitleRank, user.Level, user.CurrentEXP, user.TotalEXP,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		user.ID = id
	}
	return nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, username, password_hash, display_name, title_rank, level, current_exp, total_exp, streak_days, COALESCE(last_active_date, '') FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &u.TitleRank, &u.Level, &u.CurrentEXP, &u.TotalEXP, &u.StreakDays, &u.LastActiveDate)

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, username, display_name, title_rank, level, current_exp, total_exp, streak_days, COALESCE(last_active_date, '') FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.TitleRank, &u.Level, &u.CurrentEXP, &u.TotalEXP, &u.StreakDays, &u.LastActiveDate)

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) UpdateStats(ctx context.Context, userID int64, level, currentEXP, totalEXP int, titleRank, lastActiveDate string) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE users SET level = ?, current_exp = ?, total_exp = ?, title_rank = ?, last_active_date = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		level, currentEXP, totalEXP, titleRank, lastActiveDate, userID,
	)
	return err
}
