package services

import (
	"context"
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

const dummyPasswordHash = "$2a$10$sNzmkAV9SR7hp1iuvK4Ga.R03fZYu/iPaMNdTNtz3oGLGIzB087dG"

type AdminAuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, sessionID string) error
}

type adminAuthService struct {
	users    repositories.UserRepository
	sessions repositories.AdminSessionRepository
}

func NewAdminAuthService(users repositories.UserRepository, sessions repositories.AdminSessionRepository) AdminAuthService {
	return &adminAuthService{users: users, sessions: sessions}
}

func (s *adminAuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	hash := dummyPasswordHash
	if user != nil {
		hash = user.PasswordHash
	}
	bcryptErr := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if user == nil || user.Role != models.UserRoleAdmin || bcryptErr != nil {
		return "", apperrors.ErrInvalidCredentials
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID); err != nil {
		slog.WarnContext(ctx, "failed to update admin last login", "user_id", user.ID, "error", err)
	}
	return s.sessions.Create(ctx, user.ID)
}

func (s *adminAuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessions.Delete(ctx, sessionID)
}
