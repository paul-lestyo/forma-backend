package domain

type User struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	PasswordHash   string `json:"-"`
	DisplayName    string `json:"display_name"`
	Level          int    `json:"level"`
	CurrentEXP     int    `json:"current_exp"`
	TotalEXP       int    `json:"total_exp"`
	TitleRank      string `json:"title_rank"`
	StreakDays     int    `json:"streak_days"`
	LastActiveDate string `json:"last_active_date"`
	CreatedAt      string `json:"created_at"`
}

type HabitTemplate struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Title     string `json:"title"`
	Priority  string `json:"priority"`
	EXPReward int    `json:"exp_reward"`
	Frequency string `json:"frequency"` // "daily", "weekly", "monthly"
	IsActive  int    `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type CustomTodo struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Title     string `json:"title"`
	Priority  string `json:"priority"`
	EXPReward int    `json:"exp_reward"`
	Date      string `json:"date"` // YYYY-MM-DD
	Completed int    `json:"completed"`
	CreatedAt string `json:"created_at"`
}

type HabitLog struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	TemplateID  int64  `json:"template_id"`
	PeriodKey   string `json:"period_key"` // YYYY-MM-DD or YYYY-Www or YYYY-MM
	Completed   int    `json:"completed"`
	CompletedAt string `json:"completed_at"`
}

type QuestItem struct {
	ID         int64  `json:"id"`
	TemplateID int64  `json:"template_id,omitempty"`
	ItemType   string `json:"item_type"` // "habit" or "custom"
	Title      string `json:"title"`
	Priority   string `json:"priority"`
	EXPReward  int    `json:"exp_reward"`
	Frequency  string `json:"frequency,omitempty"`
	Date       string `json:"date"`
	Completed  bool   `json:"completed"`
}

type RegisterRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateCustomTodoRequest struct {
	Title     string `json:"title"`
	Priority  string `json:"priority"`
	EXPReward int    `json:"exp_reward"`
	Date      string `json:"date"`
}

type CreateTemplateRequest struct {
	Title     string `json:"title"`
	Priority  string `json:"priority"`
	EXPReward int    `json:"exp_reward"`
	Frequency string `json:"frequency"`
}

type ToggleQuestRequest struct {
	ID         int64  `json:"id"`
	TemplateID int64  `json:"template_id,omitempty"`
	ItemType   string `json:"item_type"`
	Date       string `json:"date"`
}

type ToggleQuestResponse struct {
	Completed  bool `json:"completed"`
	EXPGained  int  `json:"exp_gained"`
	LevelUp    bool `json:"level_up"`
	NewLevel   int  `json:"new_level"`
	CurrentEXP int  `json:"current_exp"`
	TotalEXP   int  `json:"total_exp"`
	TargetEXP  int  `json:"target_exp"`
	StreakDays int  `json:"streak_days"`
}

type RecapResponse struct {
	TotalQuestsCompleted int                  `json:"total_quests_completed"`
	StreakDays           int                  `json:"streak_days"`
	WeeklyCompletionRate float64              `json:"weekly_completion_rate"`
	TotalEXPThisWeek     int                  `json:"total_exp_this_week"`
	PeakDay              string               `json:"peak_day"`
	ActiveRoutinesCount  int                  `json:"active_routines_count"`
	RecentEXPHistory     []EXPDayHistory      `json:"exp_history"`
	WeekRange            string               `json:"week_range"`
	WeekOffset           int                  `json:"week_offset"`
	SelectedMonth        string               `json:"selected_month"`
	MonthName            string               `json:"month_name"`
	TotalEXPMonth        int                  `json:"total_exp_month"`
	MonthlyActivity      []MonthlyDayActivity `json:"monthly_activity"`
}

type EXPDayHistory struct {
	Date      string `json:"date"`
	EXPEarned int    `json:"exp_earned"`
}

type MonthlyDayActivity struct {
	Date           string `json:"date"`            // YYYY-MM-DD
	Day            int    `json:"day"`             // 1..31
	DayOfWeek      int    `json:"day_of_week"`     // 1=Mon..7=Sun
	EXPEarned      int    `json:"exp_earned"`
	Level          int    `json:"level"`           // 0..4 for heatmap shading
	CompletedCount int    `json:"completed_count"`
}
