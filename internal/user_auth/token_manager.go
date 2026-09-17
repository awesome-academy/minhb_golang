package userauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"cinema-booking/internal/models"
)

const Issuer = "cinema-booking"

type Token struct {
	Value     string
	ID        string
	ExpiresAt time.Time
}

type TokenManager struct {
	secret    []byte
	accessTTL time.Duration
	now       func() time.Time
}

func NewTokenManager(secret string, accessTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:    []byte(secret),
		accessTTL: accessTTL,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (m *TokenManager) IssueAccess(user *models.User) (Token, error) {
	id, err := newTokenID()
	if err != nil {
		return Token{}, err
	}
	now := m.now()
	expiresAt := now.Add(m.accessTTL)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Subject:   strconv.FormatInt(user.ID, 10),
			ID:        id,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Role: string(user.Role),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return Token{}, fmt.Errorf("sign access token: %w", err)
	}
	return Token{Value: signed, ID: id, ExpiresAt: expiresAt}, nil
}

func (m *TokenManager) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, m.keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(Issuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if claims.ID == "" {
		return nil, ErrInvalidToken
	}
	if _, err := claims.UserID(); err != nil {
		return nil, err
	}
	return claims, nil
}

func (m *TokenManager) keyFunc(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return m.secret, nil
}

func newTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
