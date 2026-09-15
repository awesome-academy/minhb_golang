package services

import (
	"context"
	"strings"
	"time"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/utils"
)

const TheaterPageSize = 20

type AdminTheaterService interface {
	List(ctx context.Context, search string, page int) ([]models.Theater, int64, error)
	Get(ctx context.Context, id int64) (*models.Theater, error)
	Create(ctx context.Context, form dto.AdminTheaterForm) (int64, error)
	Update(ctx context.Context, id int64, form dto.AdminTheaterForm) error
	ChangeStatus(ctx context.Context, id int64) (bool, error)
}

type adminTheaterService struct {
	theaters repositories.TheaterRepository
}

func NewAdminTheaterService(theaters repositories.TheaterRepository) AdminTheaterService {
	return &adminTheaterService{theaters: theaters}
}

func (s *adminTheaterService) List(ctx context.Context, search string, page int) ([]models.Theater, int64, error) {
	if page < 1 {
		page = 1
	}
	return s.theaters.List(ctx, strings.TrimSpace(search), (page-1)*TheaterPageSize, TheaterPageSize)
}

func (s *adminTheaterService) Get(ctx context.Context, id int64) (*models.Theater, error) {
	return s.theaters.FindByID(ctx, id)
}

func (s *adminTheaterService) Create(ctx context.Context, form dto.AdminTheaterForm) (int64, error) {
	var theater models.Theater
	applyTheaterForm(&theater, form)
	slug, err := utils.NewSlug(form.Name, "theater")
	if err != nil {
		return 0, err
	}
	theater.Slug = slug
	if err := s.theaters.Create(ctx, &theater); err != nil {
		return 0, err
	}
	return theater.ID, nil
}

func (s *adminTheaterService) Update(ctx context.Context, id int64, form dto.AdminTheaterForm) error {
	expectedUpdatedAt, err := time.Parse(time.RFC3339Nano, form.UpdatedAt)
	if err != nil {
		return apperrors.ErrRecordModified
	}
	theater, err := s.theaters.FindByID(ctx, id)
	if err != nil {
		return err
	}
	nameChanged := theater.Name != strings.TrimSpace(form.Name)
	applyTheaterForm(theater, form)
	if nameChanged {
		if theater.Slug, err = utils.NewSlug(form.Name, "theater"); err != nil {
			return err
		}
	}
	return s.theaters.Update(ctx, theater, expectedUpdatedAt)
}

func (s *adminTheaterService) ChangeStatus(ctx context.Context, id int64) (bool, error) {
	return s.theaters.ChangeStatus(ctx, id)
}

func applyTheaterForm(theater *models.Theater, form dto.AdminTheaterForm) {
	theater.Name = strings.TrimSpace(form.Name)
	theater.Address = strings.TrimSpace(form.Address)
	theater.City = strings.TrimSpace(form.City)
	theater.Phone = optionalString(form.Phone)
	theater.Description = optionalString(form.Description)
	theater.ImageURL = optionalString(form.ImageURL)
	theater.IsActive = form.IsActive
}
