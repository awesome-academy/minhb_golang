package dto

type AdminShowtimeForm struct {
	MovieID     int64    `json:"movie_id" form:"movie_id" label:"Movie" validate:"required"`
	RoomID      int64    `json:"room_id" form:"room_id" label:"Room" validate:"required"`
	StartsAt    string   `json:"starts_at" form:"starts_at" label:"Starts at" validate:"required,datetime=2006-01-02T15:04"`
	EndsAt      string   `json:"ends_at" form:"ends_at" label:"Ends at" validate:"required,datetime=2006-01-02T15:04"`
	Format      string   `json:"format" form:"format" label:"Format" validate:"required,oneof=2D 3D IMAX"`
	IsPublished bool     `json:"is_published" form:"is_published" label:"Published"`
	SeatTypeIDs []int64  `json:"seat_type_id" form:"seat_type_id" label:"Seat type" validate:"max=50,dive,gt=0"`
	Prices      []string `json:"price" form:"price" label:"Price" validate:"max=50,dive,max=20"`
	UpdatedAt   string   `json:"updated_at" form:"updated_at"`
}
