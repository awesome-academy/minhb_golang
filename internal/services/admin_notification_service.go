package services

import (
	"context"
	"encoding/json"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

type Broadcaster interface {
	Broadcast(message []byte)
}

type AdminNotificationService interface {
	BookingCreated(ctx context.Context, booking *models.Booking) error
	Recent(ctx context.Context) ([]json.RawMessage, error)
}

type adminNotificationService struct {
	notifications repositories.AdminNotificationRepository
	hub           Broadcaster
}

func NewAdminNotificationService(notifications repositories.AdminNotificationRepository, hub Broadcaster) AdminNotificationService {
	return &adminNotificationService{notifications: notifications, hub: hub}
}

func (s *adminNotificationService) BookingCreated(ctx context.Context, booking *models.Booking) error {
	payload, err := json.Marshal(dto.NewBookingNotification(booking))
	if err != nil {
		return err
	}
	pushErr := s.notifications.Push(ctx, payload)
	s.hub.Broadcast(payload)
	return pushErr
}

func (s *adminNotificationService) Recent(ctx context.Context) ([]json.RawMessage, error) {
	return s.notifications.Recent(ctx)
}
