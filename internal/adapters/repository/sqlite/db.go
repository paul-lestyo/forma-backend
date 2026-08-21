package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/glebarez/go-sqlite"
)

func InitDB() (*sql.DB, error) {
	dbDir := "./data"
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "habittracker.db")
	database, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		display_name TEXT NOT NULL,
		title_rank TEXT NOT NULL DEFAULT 'Novice Adventurer',
		level INTEGER NOT NULL DEFAULT 1,
		current_exp INTEGER NOT NULL DEFAULT 65,
		total_exp INTEGER NOT NULL DEFAULT 65,
		streak_days INTEGER NOT NULL DEFAULT 0,
		last_active_date TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS habit_templates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		priority TEXT NOT NULL CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH')),
		exp_reward INTEGER NOT NULL DEFAULT 10,
		frequency TEXT NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly')),
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS custom_todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title TEXT NOT NULL,
		priority TEXT NOT NULL CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH')),
		exp_reward INTEGER NOT NULL DEFAULT 10,
		date TEXT NOT NULL,
		completed INTEGER NOT NULL DEFAULT 0,
		completed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS habit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		template_id INTEGER NOT NULL REFERENCES habit_templates(id) ON DELETE CASCADE,
		period_key TEXT NOT NULL,
		completed INTEGER NOT NULL DEFAULT 1,
		completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, template_id, period_key)
	);

	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_custom_todos_user_date ON custom_todos(user_id, date);
	CREATE INDEX IF NOT EXISTS idx_habit_logs_user_period ON habit_logs(user_id, period_key);
	CREATE INDEX IF NOT EXISTS idx_habit_templates_user_active ON habit_templates(user_id, is_active);
	`

	if _, err := database.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to execute schema: %w", err)
	}

	return database, nil
}
