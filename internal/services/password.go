package services

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const dummyPasswordHash = "$2a$10$sNzmkAV9SR7hp1iuvK4Ga.R03fZYu/iPaMNdTNtz3oGLGIzB087dG"

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func verifyPassword(hash, password string) bool {
	if hash == "" {
		hash = dummyPasswordHash
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
