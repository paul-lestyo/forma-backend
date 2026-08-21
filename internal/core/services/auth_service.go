package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"habittracker-be/internal/core/domain"
	"habittracker-be/internal/core/ports"
)

var JWTSecret = []byte("habittracker_super_secret_jwt_key_2026")

type authService struct {
	userRepo   ports.UserRepository
	habitRepo  ports.HabitRepository
}

func NewAuthService(userRepo ports.UserRepository, habitRepo ports.HabitRepository) ports.AuthService {
	return &authService{
		userRepo:  userRepo,
		habitRepo: habitRepo,
	}
}

func generateToken(userID int64, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24 * 30).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func (s *authService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	username := strings.TrimSpace(strings.ToLower(req.Username))
	displayName := strings.TrimSpace(req.DisplayName)

	if username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}

	if displayName == "" {
		displayName = username
	}

	existingUser, _ := s.userRepo.FindByUsername(ctx, username)
	if existingUser != nil {
		return nil, errors.New("username already taken")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		DisplayName:  displayName,
		TitleRank:    "Novice Adventurer",
		Level:        1,
		CurrentEXP:   65,
		TotalEXP:     65,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Seed default habit templates matching mockup
	seedTemplates := []struct {
		Title     string
		Priority  string
		EXPReward int
		Frequency string
	}{
		{"Review pull request & monitoring log shift", "HIGH", 25, "daily"},
		{"Latihan 10 kosakata baru", "MEDIUM", 15, "daily"},
		{"Minum air 2 liter & stretching", "LOW", 10, "daily"},
		{"Backup database & cleanup docker cache", "HIGH", 60, "monthly"},
		{"Review progres investasi & alokasi portofolio", "MEDIUM", 40, "weekly"},
	}

	for _, tmpl := range seedTemplates {
		_ = s.habitRepo.Create(ctx, &domain.HabitTemplate{
			UserID:    user.ID,
			Title:     tmpl.Title,
			Priority:  tmpl.Priority,
			EXPReward: tmpl.EXPReward,
			Frequency: tmpl.Frequency,
			IsActive:  1,
		})
	}

	token, err := generateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *authService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	username := strings.TrimSpace(strings.ToLower(req.Username))

	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil || user == nil {
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	token, err := generateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *authService) GetProfile(ctx context.Context, userID int64) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}
