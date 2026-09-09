package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID              int64  `gorm:"primaryKey"`
	Email           string `gorm:"not null"`
	PasswordHash    string `gorm:"not null"`
	FullName        string `gorm:"not null"`
	Phone           *string
	DateOfBirth     *time.Time `gorm:"type:date"`
	AvatarURL       *string
	Role            UserRole `gorm:"type:user_role;not null;default:'user'"`
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	DeletedAt       gorm.DeletedAt
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
