package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/user_auth"
)

const dateOfBirthLayout = time.DateOnly

type RegisterUserInput struct {
	Email       string
	Password    string
	FullName    string
	Phone       string
	DateOfBirth string
	AvatarURL   string
}

type UserAuthService interface {
	Register(ctx context.Context, input RegisterUserInput) (*models.User, string, error)
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, claims *userauth.Claims) error
}

type userAuthService struct {
	users  repositories.UserRepository
	tokens repositories.UserTokenRepository
	jwt    *userauth.TokenManager
	now    func() time.Time
}

func NewUserAuthService(users repositories.UserRepository, tokens repositories.UserTokenRepository, jwt *userauth.TokenManager) UserAuthService {
	return &userAuthService{
		users:  users,
		tokens: tokens,
		jwt:    jwt,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (s *userAuthService) Register(ctx context.Context, input RegisterUserInput) (*models.User, string, error) {
	dateOfBirth, err := parseDateOfBirth(input.DateOfBirth)
	if err != nil {
		return nil, "", err
	}
	hash, err := hashPassword(input.Password)
	if err != nil {
		return nil, "", err
	}
	user := &models.User{
		Email:        input.Email,
		PasswordHash: hash,
		FullName:     input.FullName,
		DateOfBirth:  dateOfBirth,
		Role:         models.UserRoleUser,
	}
	if input.Phone != "" {
		phone := input.Phone
		user.Phone = &phone
	}
	if input.AvatarURL != "" {
		avatarURL := input.AvatarURL
		user.AvatarURL = &avatarURL
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.jwt.IssueAccess(user)
	if err != nil {
		return nil, "", err
	}
	return user, token.Value, nil
}

func (s *userAuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	hash := ""
	if user != nil {
		hash = user.PasswordHash
	}
	passwordOK := verifyPassword(hash, password)
	if user == nil || !passwordOK {
		return "", apperrors.ErrInvalidCredentials
	}

	if err := s.users.UpdateLastLogin(ctx, user.ID); err != nil {
		slog.WarnContext(ctx, "failed to update user last login", "user_id", user.ID, "error", err)
	}

	token, err := s.jwt.IssueAccess(user)
	if err != nil {
		return "", err
	}
	return token.Value, nil
}

func parseDateOfBirth(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	dateOfBirth, err := time.ParseInLocation(dateOfBirthLayout, value, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("parse dateOfBirth: %w", err)
	}
	return &dateOfBirth, nil
}

func (s *userAuthService) Logout(ctx context.Context, claims *userauth.Claims) error {
	if claims == nil || claims.ExpiresAt == nil {
		return errors.New("logout: access token claims are missing")
	}
	remaining := claims.ExpiresAt.Time.Sub(s.now())
	if err := s.tokens.RevokeAccess(ctx, claims.ID, remaining); err != nil {
		return fmt.Errorf("revoke access token: %w", err)
	}
	return nil
}
