package repositories

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

const (
	adminNotificationsKey = "admin_notifications"
	adminNotificationsMax = 100
)

type AdminNotificationRepository interface {
	Push(ctx context.Context, payload []byte) error
	Recent(ctx context.Context) ([]json.RawMessage, error)
}

type adminNotificationRepository struct {
	client *redis.Client
}

func NewAdminNotificationRepository(client *redis.Client) AdminNotificationRepository {
	return &adminNotificationRepository{client: client}
}

func (r *adminNotificationRepository) Push(ctx context.Context, payload []byte) error {
	pipe := r.client.TxPipeline()
	pipe.LPush(ctx, adminNotificationsKey, payload)
	pipe.LTrim(ctx, adminNotificationsKey, 0, adminNotificationsMax-1)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *adminNotificationRepository) Recent(ctx context.Context) ([]json.RawMessage, error) {
	values, err := r.client.LRange(ctx, adminNotificationsKey, 0, adminNotificationsMax-1).Result()
	if err != nil {
		return nil, err
	}
	items := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		items = append(items, json.RawMessage(value))
	}
	return items, nil
}
