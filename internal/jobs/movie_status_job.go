package jobs

import (
	"context"
	"log/slog"
	"time"

	"cinema-booking/internal/repositories"
	"cinema-booking/internal/utils"
)

type MovieStatusJob struct {
	movies repositories.MovieRepository
}

func NewMovieStatusJob(movies repositories.MovieRepository) *MovieStatusJob {
	return &MovieStatusJob{movies: movies}
}

func (j *MovieStatusJob) Run(ctx context.Context) {
	from, to := utils.DayRange(time.Now())
	promoted, err := j.movies.PromoteNowShowing(ctx, from, to)
	if err != nil {
		if ctx.Err() == nil {
			slog.ErrorContext(ctx, "promote movies to now_showing", "error", err)
		}
		return
	}
	if promoted > 0 {
		slog.InfoContext(ctx, "promoted movies to now_showing", "movies", promoted, "day", from.Format(time.DateOnly))
	}
}
