package services

import (
	"context"
	"strings"
	"time"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

type RoomWithSeats struct {
	models.Room
	SeatCount repositories.SeatCount
}

type AdminRoomService interface {
	ListByTheater(ctx context.Context, theaterID int64) (*models.Theater, []RoomWithSeats, error)
	GetTheater(ctx context.Context, theaterID int64) (*models.Theater, error)
	Get(ctx context.Context, id int64) (*models.Room, error)
	Create(ctx context.Context, theaterID int64, form dto.AdminRoomForm) (int64, error)
	Update(ctx context.Context, id int64, form dto.AdminRoomForm) error
	ChangeStatus(ctx context.Context, id int64) (bool, error)
}

type adminRoomService struct {
	theaters repositories.TheaterRepository
	rooms    repositories.RoomRepository
	seats    repositories.SeatRepository
}

func NewAdminRoomService(theaters repositories.TheaterRepository, rooms repositories.RoomRepository, seats repositories.SeatRepository) AdminRoomService {
	return &adminRoomService{theaters: theaters, rooms: rooms, seats: seats}
}

func (s *adminRoomService) ListByTheater(ctx context.Context, theaterID int64) (*models.Theater, []RoomWithSeats, error) {
	theater, err := s.theaters.FindByID(ctx, theaterID)
	if err != nil {
		return nil, nil, err
	}
	rooms, err := s.rooms.ListByTheater(ctx, theaterID)
	if err != nil {
		return nil, nil, err
	}
	ids := make([]int64, 0, len(rooms))
	for _, room := range rooms {
		ids = append(ids, room.ID)
	}
	counts, err := s.seats.SeatCounts(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	result := make([]RoomWithSeats, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, RoomWithSeats{Room: room, SeatCount: counts[room.ID]})
	}
	return theater, result, nil
}

func (s *adminRoomService) GetTheater(ctx context.Context, theaterID int64) (*models.Theater, error) {
	return s.theaters.FindByID(ctx, theaterID)
}

func (s *adminRoomService) Get(ctx context.Context, id int64) (*models.Room, error) {
	return s.rooms.FindByID(ctx, id)
}

func (s *adminRoomService) Create(ctx context.Context, theaterID int64, form dto.AdminRoomForm) (int64, error) {
	if _, err := s.theaters.FindByID(ctx, theaterID); err != nil {
		return 0, err
	}
	room := models.Room{TheaterID: theaterID}
	applyRoomForm(&room, form)
	if err := s.ensureNameFree(ctx, theaterID, room.Name, 0); err != nil {
		return 0, err
	}
	if err := s.rooms.Create(ctx, &room); err != nil {
		return 0, err
	}
	return room.ID, nil
}

func (s *adminRoomService) Update(ctx context.Context, id int64, form dto.AdminRoomForm) error {
	expectedUpdatedAt, err := time.Parse(time.RFC3339Nano, form.UpdatedAt)
	if err != nil {
		return apperrors.ErrRecordTokenInvalid
	}
	room, err := s.rooms.FindByID(ctx, id)
	if err != nil {
		return err
	}
	applyRoomForm(room, form)
	if err := s.ensureNameFree(ctx, room.TheaterID, room.Name, room.ID); err != nil {
		return err
	}
	return s.rooms.Update(ctx, room, expectedUpdatedAt)
}

func (s *adminRoomService) ChangeStatus(ctx context.Context, id int64) (bool, error) {
	return s.rooms.ChangeStatus(ctx, id)
}

func (s *adminRoomService) ensureNameFree(ctx context.Context, theaterID int64, name string, excludeID int64) error {
	taken, err := s.rooms.NameExists(ctx, theaterID, name, excludeID)
	if err != nil {
		return err
	}
	if taken {
		return apperrors.ErrRoomNameTaken
	}
	return nil
}

func applyRoomForm(room *models.Room, form dto.AdminRoomForm) {
	room.Name = strings.TrimSpace(form.Name)
	room.IsActive = form.IsActive
}
