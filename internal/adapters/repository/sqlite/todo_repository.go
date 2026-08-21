package sqlite

import (
	"context"
	"database/sql"
	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

type todoRepo struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) ports.TodoRepository {
	return &todoRepo{db: db}
}

func (r *todoRepo) Create(ctx context.Context, todo *domain.CustomTodo) error {
	res, err := r.db.ExecContext(
		ctx,
		"INSERT INTO custom_todos (user_id, title, priority, exp_reward, date, completed) VALUES (?, ?, ?, ?, ?, 0)",
		todo.UserID, todo.Title, todo.Priority, todo.EXPReward, todo.Date,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		todo.ID = id
		todo.Completed = 0
	}
	return nil
}

func (r *todoRepo) FindByUserIDAndDate(ctx context.Context, userID int64, date string) ([]domain.CustomTodo, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, user_id, title, priority, exp_reward, date, completed, created_at FROM custom_todos WHERE user_id = ? AND date = ?",
		userID, date,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []domain.CustomTodo
	for rows.Next() {
		var todo domain.CustomTodo
		if err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Priority, &todo.EXPReward, &todo.Date, &todo.Completed, &todo.CreatedAt); err != nil {
			continue
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func (r *todoRepo) FindByID(ctx context.Context, id int64, userID int64) (*domain.CustomTodo, error) {
	var todo domain.CustomTodo
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, user_id, title, priority, exp_reward, date, completed FROM custom_todos WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Priority, &todo.EXPReward, &todo.Date, &todo.Completed)

	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (r *todoRepo) UpdateCompletion(ctx context.Context, id int64, completed int) error {
	var query string
	if completed == 1 {
		query = "UPDATE custom_todos SET completed = 1, completed_at = CURRENT_TIMESTAMP WHERE id = ?"
	} else {
		query = "UPDATE custom_todos SET completed = 0, completed_at = NULL WHERE id = ?"
	}
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *todoRepo) Delete(ctx context.Context, id int64, userID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM custom_todos WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func (r *todoRepo) CountCompletedByUserID(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM custom_todos WHERE user_id = ? AND completed = 1", userID).Scan(&count)
	return count, err
}

func (r *todoRepo) SumEXPByUserIDAndDate(ctx context.Context, userID int64, date string) (int, error) {
	var sum sql.NullInt64
	err := r.db.QueryRowContext(ctx, "SELECT SUM(exp_reward) FROM custom_todos WHERE user_id = ? AND date = ? AND completed = 1", userID, date).Scan(&sum)
	if err != nil {
		return 0, err
	}
	if sum.Valid {
		return int(sum.Int64), nil
	}
	return 0, nil
}
