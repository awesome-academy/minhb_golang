package dto

type AdminMovieForm struct {
	Title         string   `json:"title" form:"title" label:"Title" validate:"required,max=200"`
	OriginalTitle string   `json:"original_title" form:"original_title" label:"Original title" validate:"max=200"`
	Description   string   `json:"description" form:"description" label:"Description" validate:"max=5000"`
	DurationMin   int      `json:"duration_min" form:"duration_min" label:"Duration (minutes)" validate:"required,min=1,max=600"`
	AgeRating     string   `json:"age_rating" form:"age_rating" label:"Age rating" validate:"required,oneof=P K T13 T16 T18 C"`
	Language      string   `json:"language" form:"language" label:"Language" validate:"max=100"`
	Director      string   `json:"director" form:"director" label:"Director" validate:"max=200"`
	CastNames     []string `json:"cast_name" form:"cast_name" label:"Actor name" validate:"max=50,dive,max=200"`
	CastRoles     []string `json:"cast_role" form:"cast_role" label:"Character name" validate:"max=50,dive,max=200"`
	PosterURL     string   `json:"poster_url" form:"poster_url" label:"Poster URL" validate:"omitempty,url,max=2048"`
	BackdropURL   string   `json:"backdrop_url" form:"backdrop_url" label:"Backdrop URL" validate:"omitempty,url,max=2048"`
	TrailerURL    string   `json:"trailer_url" form:"trailer_url" label:"Trailer URL" validate:"omitempty,url,max=2048"`
	ReleaseDate   string   `json:"release_date" form:"release_date" label:"Release date" validate:"required,datetime=2006-01-02"`
	Status        string   `json:"status" form:"status" label:"Status" validate:"required,oneof=coming_soon now_showing ended"`
	GenreIDs      []int64  `json:"genre_ids" form:"genre_ids" label:"Genres" validate:"max=50,dive,gt=0"`
	UpdatedAt     string   `json:"updated_at" form:"updated_at"`
}
