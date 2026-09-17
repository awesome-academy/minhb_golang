package services

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

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

	hash := ""
	if user != nil {
		hash = user.PasswordHash
	}
	passwordOK := verifyPassword(hash, password)
	if user == nil || user.Role != models.UserRoleAdmin || !passwordOK {
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
