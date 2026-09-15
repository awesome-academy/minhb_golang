package services

import (
	"context"
	"slices"
	"strings"
	"time"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/utils"
)

const (
	MoviePageSize     = 20
	releaseDateLayout = "2006-01-02"
)

type AdminMovieService interface {
	List(ctx context.Context, search string, page int) ([]models.Movie, int64, error)
	Get(ctx context.Context, id int64) (*models.Movie, error)
	Genres(ctx context.Context) ([]models.Genre, error)
	Create(ctx context.Context, form dto.AdminMovieForm) (int64, error)
	Update(ctx context.Context, id int64, form dto.AdminMovieForm) error
	Delete(ctx context.Context, id int64) error
}

type adminMovieService struct {
	movies repositories.MovieRepository
	genres repositories.GenreRepository
}

func NewAdminMovieService(movies repositories.MovieRepository, genres repositories.GenreRepository) AdminMovieService {
	return &adminMovieService{movies: movies, genres: genres}
}

func (s *adminMovieService) List(ctx context.Context, search string, page int) ([]models.Movie, int64, error) {
	if page < 1 {
		page = 1
	}
	return s.movies.List(ctx, strings.TrimSpace(search), (page-1)*MoviePageSize, MoviePageSize)
}

func (s *adminMovieService) Get(ctx context.Context, id int64) (*models.Movie, error) {
	return s.movies.FindByID(ctx, id)
}

func (s *adminMovieService) Genres(ctx context.Context) ([]models.Genre, error) {
	return s.genres.ListAll(ctx)
}

func (s *adminMovieService) Create(ctx context.Context, form dto.AdminMovieForm) (int64, error) {
	var movie models.Movie
	if err := applyMovieForm(&movie, form); err != nil {
		return 0, err
	}
	genreIDs, err := s.checkGenres(ctx, form.GenreIDs)
	if err != nil {
		return 0, err
	}
	if movie.Slug, err = utils.NewSlug(form.Title, "movie"); err != nil {
		return 0, err
	}
	if err := s.movies.Create(ctx, &movie, genreIDs); err != nil {
		return 0, err
	}
	return movie.ID, nil
}

func (s *adminMovieService) Update(ctx context.Context, id int64, form dto.AdminMovieForm) error {
	expectedUpdatedAt, err := time.Parse(time.RFC3339Nano, form.UpdatedAt)
	if err != nil {
		return apperrors.ErrRecordTokenInvalid
	}
	movie, err := s.movies.FindByID(ctx, id)
	if err != nil {
		return err
	}
	titleChanged := movie.Title != strings.TrimSpace(form.Title)
	if err := applyMovieForm(movie, form); err != nil {
		return err
	}
	genreIDs, err := s.checkGenres(ctx, form.GenreIDs)
	if err != nil {
		return err
	}
	if titleChanged {
		if movie.Slug, err = utils.NewSlug(form.Title, "movie"); err != nil {
			return err
		}
	}
	return s.movies.Update(ctx, movie, genreIDs, expectedUpdatedAt)
}

func (s *adminMovieService) Delete(ctx context.Context, id int64) error {
	return s.movies.Delete(ctx, id)
}

func (s *adminMovieService) checkGenres(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	ids = slices.Clone(ids)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	count, err := s.genres.CountByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	if count != int64(len(ids)) {
		return nil, apperrors.ErrMovieGenreInvalid
	}
	return ids, nil
}

func applyMovieForm(movie *models.Movie, form dto.AdminMovieForm) error {
	cast, err := parseCastMembers(form.CastNames, form.CastRoles)
	if err != nil {
		return err
	}
	releaseDate, err := time.Parse(releaseDateLayout, form.ReleaseDate)
	if err != nil {
		return err
	}
	movie.Title = strings.TrimSpace(form.Title)
	movie.OriginalTitle = optionalString(form.OriginalTitle)
	movie.Description = optionalString(form.Description)
	movie.DurationMin = form.DurationMin
	movie.AgeRating = models.AgeRating(form.AgeRating)
	movie.Language = optionalString(form.Language)
	movie.Director = optionalString(form.Director)
	movie.CastMembers = cast
	movie.PosterURL = optionalString(form.PosterURL)
	movie.BackdropURL = optionalString(form.BackdropURL)
	movie.TrailerURL = optionalString(form.TrailerURL)
	movie.ReleaseDate = releaseDate
	movie.Status = models.MovieStatus(form.Status)
	return nil
}

func parseCastMembers(names, roles []string) (models.CastMembers, error) {
	var cast models.CastMembers
	for i := 0; i < max(len(names), len(roles)); i++ {
		name := strings.TrimSpace(at(names, i))
		role := strings.TrimSpace(at(roles, i))
		if name == "" && role == "" {
			continue
		}
		if name == "" {
			return nil, apperrors.ErrMovieCastInvalid
		}
		cast = append(cast, models.CastMember{Name: name, Role: role})
	}
	return cast, nil
}

func at(values []string, i int) string {
	if i < len(values) {
		return values[i]
	}
	return ""
}

func optionalString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
