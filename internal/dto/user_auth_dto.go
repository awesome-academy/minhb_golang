package dto

import (
	"time"

	"cinema-booking/internal/models"
)

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	Password    string `json:"password" validate:"required,min=8,max=72" example:"Password123"`
	FullName    string `json:"fullName" validate:"required,max=100" example:"Nguyen Van A"`
	Phone       string `json:"phone" validate:"omitempty,max=20" example:"0901234567"`
	DateOfBirth string `json:"dateOfBirth" validate:"omitempty,datetime=2006-01-02" example:"1995-06-15"`
	AvatarURL   string `json:"avatarUrl" validate:"omitempty,url,max=2048" example:"https://cdn.example.com/avatars/1.png"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"Password123"`
}

type UserResponse struct {
	ID          int64     `json:"id" example:"1"`
	Email       string    `json:"email" example:"user@example.com"`
	FullName    string    `json:"fullName" example:"Nguyen Van A"`
	Phone       *string   `json:"phone" example:"0901234567"`
	DateOfBirth *string   `json:"dateOfBirth" example:"1995-06-15"`
	AvatarURL   *string   `json:"avatarUrl" example:"https://cdn.example.com/avatars/1.png"`
	Role        string    `json:"role" example:"user"`
	CreatedAt   time.Time `json:"createdAt" example:"2026-09-17T08:00:00Z"`
}

type TokenResponse struct {
	AccessToken string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

func NewUserResponse(user *models.User) UserResponse {
	var dateOfBirth *string
	if user.DateOfBirth != nil {
		formatted := user.DateOfBirth.Format(time.DateOnly)
		dateOfBirth = &formatted
	}
	return UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		Phone:       user.Phone,
		DateOfBirth: dateOfBirth,
		AvatarURL:   user.AvatarURL,
		Role:        string(user.Role),
		CreatedAt:   user.CreatedAt,
	}
}

func NewTokenResponse(accessToken string) TokenResponse {
	return TokenResponse{AccessToken: accessToken}
}
