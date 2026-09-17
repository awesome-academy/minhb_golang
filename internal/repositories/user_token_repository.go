package repositories

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const userAccessRevokedKeyPrefix = "user_access_revoked:"

type UserTokenRepository interface {
	RevokeAccess(ctx context.Context, jti string, ttl time.Duration) error
	IsAccessRevoked(ctx context.Context, jti string) (bool, error)
}

type userTokenRepository struct {
	client *redis.Client
}

func NewUserTokenRepository(client *redis.Client) UserTokenRepository {
	return &userTokenRepository{client: client}
}

func (r *userTokenRepository) RevokeAccess(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return r.client.Set(ctx, userAccessRevokedKeyPrefix+jti, 1, ttl).Err()
}

func (r *userTokenRepository) IsAccessRevoked(ctx context.Context, jti string) (bool, error) {
	count, err := r.client.Exists(ctx, userAccessRevokedKeyPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
