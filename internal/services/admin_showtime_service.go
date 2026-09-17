package services

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/utils"
)

const (
	ShowtimeInputLayout = "2006-01-02T15:04"
	ShowtimePageSize    = 20
)

type ShowtimeList struct {
	Theaters      []models.Theater
	Showtimes     []models.Showtime
	BookingCounts map[int64]int64
	Total         int64
}

type ShowtimeFormData struct {
	Movies    []models.Movie
	Rooms     []models.Room
	SeatTypes []models.SeatType
	Prefill   map[int64]decimal.Decimal
}

type ShowtimeDetail struct {
	Showtime    *models.Showtime
	HasBookings bool
}

type AdminShowtimeService interface {
	List(ctx context.Context, filter repositories.ShowtimeFilter, page int) (*ShowtimeList, error)
	FormData(ctx context.Context, theaterID int64) (*ShowtimeFormData, error)
	Get(ctx context.Context, id int64) (*ShowtimeDetail, error)
	Create(ctx context.Context, form dto.AdminShowtimeForm) (*models.Showtime, error)
	Update(ctx context.Context, id int64, form dto.AdminShowtimeForm) (*models.Showtime, error)
	ChangePublished(ctx context.Context, id int64) (bool, error)
	Cancel(ctx context.Context, id int64) (int64, error)
}

type adminShowtimeService struct {
	theaters  repositories.TheaterRepository
	rooms     repositories.RoomRepository
	seats     repositories.SeatRepository
	seatTypes repositories.SeatTypeRepository
	movies    repositories.MovieRepository
	showtimes repositories.ShowtimeRepository
}

func NewAdminShowtimeService(
	theaters repositories.TheaterRepository,
	rooms repositories.RoomRepository,
	seats repositories.SeatRepository,
	seatTypes repositories.SeatTypeRepository,
	movies repositories.MovieRepository,
	showtimes repositories.ShowtimeRepository,
) AdminShowtimeService {
	return &adminShowtimeService{theaters: theaters, rooms: rooms, seats: seats, seatTypes: seatTypes, movies: movies, showtimes: showtimes}
}

func (s *adminShowtimeService) List(ctx context.Context, filter repositories.ShowtimeFilter, page int) (*ShowtimeList, error) {
	theaters, err := s.theaters.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	if filter.TheaterID > 0 && !slices.ContainsFunc(theaters, func(theater models.Theater) bool { return theater.ID == filter.TheaterID }) {
		return nil, gorm.ErrRecordNotFound
	}
	if page < 1 {
		page = 1
	}
	showtimes, total, err := s.showtimes.List(ctx, filter, (page-1)*ShowtimePageSize, ShowtimePageSize)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(showtimes))
	for _, showtime := range showtimes {
		ids = append(ids, showtime.ID)
	}
	counts, err := s.showtimes.ActiveBookingCounts(ctx, ids)
	if err != nil {
		return nil, err
	}
	return &ShowtimeList{Theaters: theaters, Showtimes: showtimes, BookingCounts: counts, Total: total}, nil
}

func (s *adminShowtimeService) FormData(ctx context.Context, theaterID int64) (*ShowtimeFormData, error) {
	movies, err := s.movies.ListAvailable(ctx)
	if err != nil {
		return nil, err
	}
	rooms, err := s.rooms.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	seatTypes, err := s.seatTypes.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	data := &ShowtimeFormData{Movies: movies, Rooms: rooms, SeatTypes: seatTypes, Prefill: map[int64]decimal.Decimal{}}
	if theaterID <= 0 {
		return data, nil
	}
	latest, err := s.showtimes.LatestInTheater(ctx, theaterID)
	if err != nil {
		return nil, err
	}
	if latest != nil {
		for _, price := range latest.Prices {
			data.Prefill[price.SeatTypeID] = price.Price
		}
	}
	return data, nil
}

func (s *adminShowtimeService) Get(ctx context.Context, id int64) (*ShowtimeDetail, error) {
	showtime, err := s.showtimes.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	hasBookings, err := s.showtimes.HasActiveBookings(ctx, id)
	if err != nil {
		return nil, err
	}
	return &ShowtimeDetail{Showtime: showtime, HasBookings: hasBookings}, nil
}

func (s *adminShowtimeService) Create(ctx context.Context, form dto.AdminShowtimeForm) (*models.Showtime, error) {
	var showtime models.Showtime
	room, prices, err := s.applyShowtimeForm(ctx, &showtime, form)
	if err != nil {
		return nil, err
	}
	showtime.IsPublished = false
	if err := s.showtimes.Create(ctx, &showtime, prices); err != nil {
		return nil, err
	}
	showtime.Room = *room
	return &showtime, nil
}

func (s *adminShowtimeService) Update(ctx context.Context, id int64, form dto.AdminShowtimeForm) (*models.Showtime, error) {
	expectedUpdatedAt, err := time.Parse(time.RFC3339Nano, form.UpdatedAt)
	if err != nil {
		return nil, apperrors.ErrRecordTokenInvalid
	}
	showtime, err := s.showtimes.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	room, prices, err := s.applyShowtimeForm(ctx, showtime, form)
	if err != nil {
		return nil, err
	}
	showtime.IsPublished = form.IsPublished
	if err := s.showtimes.Update(ctx, showtime, prices, expectedUpdatedAt); err != nil {
		return nil, err
	}
	showtime.Room = *room
	return showtime, nil
}

func (s *adminShowtimeService) ChangePublished(ctx context.Context, id int64) (bool, error) {
	return s.showtimes.ChangePublished(ctx, id)
}

func (s *adminShowtimeService) Cancel(ctx context.Context, id int64) (int64, error) {
	return s.showtimes.Cancel(ctx, id)
}

func (s *adminShowtimeService) applyShowtimeForm(ctx context.Context, showtime *models.Showtime, form dto.AdminShowtimeForm) (*models.Room, []models.ShowtimePrice, error) {
	startsAt, err := time.ParseInLocation(ShowtimeInputLayout, form.StartsAt, utils.Location)
	if err != nil {
		return nil, nil, err
	}
	endsAt, err := time.ParseInLocation(ShowtimeInputLayout, form.EndsAt, utils.Location)
	if err != nil {
		return nil, nil, err
	}
	keepMovie := showtime.MovieID == form.MovieID
	keepRoom := showtime.RoomID == form.RoomID

	movie, err := s.movies.FindByID(ctx, form.MovieID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !keepMovie && movie.Status == models.MovieStatusEnded) {
		return nil, nil, apperrors.ErrShowtimeMovieInvalid
	}
	if err != nil {
		return nil, nil, err
	}
	room, err := s.rooms.FindByID(ctx, form.RoomID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && !keepRoom && !room.IsActive) {
		return nil, nil, apperrors.ErrShowtimeRoomInvalid
	}
	if err != nil {
		return nil, nil, err
	}
	if !keepRoom {
		counts, err := s.seats.SeatCounts(ctx, []int64{room.ID})
		if err != nil {
			return nil, nil, err
		}
		if counts[room.ID].Total == 0 {
			return nil, nil, apperrors.ErrRoomHasNoSeats
		}
	}
	if !startsAt.After(time.Now()) {
		return nil, nil, apperrors.ErrShowtimeInPast
	}
	if endsAt.Before(startsAt.Add(time.Duration(movie.DurationMin) * time.Minute)) {
		return nil, nil, apperrors.ErrShowtimeEndsTooEarly
	}
	seatTypes, err := s.seatTypes.ListAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	prices, err := parsePrices(form, seatTypes)
	if err != nil {
		return nil, nil, err
	}
	showtime.MovieID = movie.ID
	showtime.RoomID = room.ID
	showtime.StartsAt = startsAt
	showtime.EndsAt = endsAt
	showtime.Format = models.ShowtimeFormat(form.Format)
	return room, prices, nil
}

var (
	showtimeMaxPrice = decimal.NewFromInt(10_000_000_000)
	priceFormat      = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)
)

func parsePrices(form dto.AdminShowtimeForm, seatTypes []models.SeatType) ([]models.ShowtimePrice, error) {
	values := make(map[int64]string, len(form.SeatTypeIDs))
	for i, seatTypeID := range form.SeatTypeIDs {
		values[seatTypeID] = strings.TrimSpace(at(form.Prices, i))
	}
	fieldErrors := apperrors.FieldErrors{}
	prices := make([]models.ShowtimePrice, 0, len(seatTypes))
	for _, seatType := range seatTypes {
		price, message := parsePrice(values[seatType.ID], seatType.Name)
		if message != "" {
			fieldErrors["price_"+strconv.FormatInt(seatType.ID, 10)] = message
			continue
		}
		prices = append(prices, models.ShowtimePrice{SeatTypeID: seatType.ID, Price: price})
	}
	if len(fieldErrors) > 0 {
		return nil, fieldErrors
	}
	return prices, nil
}

func parsePrice(raw, seatTypeName string) (decimal.Decimal, string) {
	label := "Price for " + seatTypeName
	if raw == "" {
		return decimal.Zero, label + " is required"
	}
	if !priceFormat.MatchString(raw) {
		return decimal.Zero, label + " must be a number"
	}
	price, err := decimal.NewFromString(raw)
	switch {
	case err != nil:
		return decimal.Zero, label + " must be a number"
	case !price.Equal(price.Round(2)):
		return decimal.Zero, label + " must have at most 2 decimals"
	case price.IsNegative():
		return decimal.Zero, label + " must be at least 0"
	case price.GreaterThanOrEqual(showtimeMaxPrice):
		return decimal.Zero, label + " is too large"
	}
	return price, ""
}
