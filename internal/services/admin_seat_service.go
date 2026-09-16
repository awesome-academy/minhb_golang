package services

import (
	"context"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

type SeatMap struct {
	Room         *models.Room
	Seats        []models.Seat
	SeatTypes    []models.SeatType
	HasShowtimes bool
}

type AdminSeatService interface {
	Map(ctx context.Context, roomID int64) (*SeatMap, error)
	Generate(ctx context.Context, roomID int64, form dto.AdminSeatGenerateForm) (int, error)
	ChangeRowTypes(ctx context.Context, roomID int64, form dto.AdminSeatRowTypesForm) (int, error)
	ChangeStatus(ctx context.Context, roomID, seatID int64) (*models.Seat, error)
}

type adminSeatService struct {
	rooms     repositories.RoomRepository
	seats     repositories.SeatRepository
	seatTypes repositories.SeatTypeRepository
}

func NewAdminSeatService(rooms repositories.RoomRepository, seats repositories.SeatRepository, seatTypes repositories.SeatTypeRepository) AdminSeatService {
	return &adminSeatService{rooms: rooms, seats: seats, seatTypes: seatTypes}
}

func (s *adminSeatService) Map(ctx context.Context, roomID int64) (*SeatMap, error) {
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	seats, err := s.seats.ListByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	seatTypes, err := s.seatTypes.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	hasShowtimes, err := s.seats.HasShowtimes(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return &SeatMap{Room: room, Seats: seats, SeatTypes: seatTypes, HasShowtimes: hasShowtimes}, nil
}

func (s *adminSeatService) Generate(ctx context.Context, roomID int64, form dto.AdminSeatGenerateForm) (int, error) {
	valid, err := s.seatTypeIDs(ctx)
	if err != nil {
		return 0, err
	}
	if !valid[form.SeatTypeID] {
		return 0, apperrors.ErrSeatTypeInvalid
	}
	return s.seats.Regenerate(ctx, roomID, form.Rows, form.SeatsPerRow, form.SeatTypeID)
}

func (s *adminSeatService) ChangeRowTypes(ctx context.Context, roomID int64, form dto.AdminSeatRowTypesForm) (int, error) {
	if len(form.RowLabels) != len(form.SeatTypeIDs) {
		return 0, apperrors.ErrRowTypesMismatch
	}
	if _, err := s.rooms.FindByID(ctx, roomID); err != nil {
		return 0, err
	}
	valid, err := s.seatTypeIDs(ctx)
	if err != nil {
		return 0, err
	}
	rowTypes := make([]repositories.RowSeatType, 0, len(form.RowLabels))
	for i, label := range form.RowLabels {
		if !valid[form.SeatTypeIDs[i]] {
			return 0, apperrors.ErrSeatTypeInvalid
		}
		rowTypes = append(rowTypes, repositories.RowSeatType{RowLabel: label, SeatTypeID: form.SeatTypeIDs[i]})
	}
	return s.seats.ChangeRowTypes(ctx, roomID, rowTypes)
}

func (s *adminSeatService) ChangeStatus(ctx context.Context, roomID, seatID int64) (*models.Seat, error) {
	return s.seats.ChangeStatus(ctx, roomID, seatID)
}

func (s *adminSeatService) seatTypeIDs(ctx context.Context) (map[int64]bool, error) {
	seatTypes, err := s.seatTypes.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	ids := make(map[int64]bool, len(seatTypes))
	for _, seatType := range seatTypes {
		ids[seatType.ID] = true
	}
	return ids, nil
}
