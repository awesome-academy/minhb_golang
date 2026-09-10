package repositories

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	apperrors "cinema-booking/internal/errors"
)

const adminSessionKeyPrefix = "admin_session:"

type AdminSessionRepository interface {
	Create(ctx context.Context, userID int64) (string, error)
	FindUserID(ctx context.Context, id string) (int64, error)
	Delete(ctx context.Context, id string) error
}

type adminSessionRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewAdminSessionRepository(client *redis.Client, ttl time.Duration) AdminSessionRepository {
	return &adminSessionRepository{client: client, ttl: ttl}
}

func (r *adminSessionRepository) Create(ctx context.Context, userID int64) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	id := base64.RawURLEncoding.EncodeToString(buf)
	if err := r.client.Set(ctx, adminSessionKeyPrefix+id, userID, r.ttl).Err(); err != nil {
		return "", err
	}
	return id, nil
}

func (r *adminSessionRepository) FindUserID(ctx context.Context, id string) (int64, error) {
	userID, err := r.client.Get(ctx, adminSessionKeyPrefix+id).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, apperrors.ErrSessionNotFound
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *adminSessionRepository) Delete(ctx context.Context, id string) error {
	return r.client.Del(ctx, adminSessionKeyPrefix+id).Err()
}
