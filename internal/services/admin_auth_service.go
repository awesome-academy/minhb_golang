package services

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", apperrors.ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}
	if user.Role != models.UserRoleAdmin {
		return "", apperrors.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", apperrors.ErrInvalidCredentials
	}
	if err := s.users.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", err
	}
	return s.sessions.Create(ctx, user.ID)
}

func (s *adminAuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessions.Delete(ctx, sessionID)
}
