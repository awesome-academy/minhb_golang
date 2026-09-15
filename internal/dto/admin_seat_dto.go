package dto

type AdminSeatGenerateForm struct {
	Rows        int   `json:"rows" form:"rows" label:"Rows" validate:"min=1,max=702"`
	SeatsPerRow int   `json:"seats_per_row" form:"seats_per_row" label:"Seats per row" validate:"min=1,max=100"`
	SeatTypeID  int64 `json:"seat_type_id" form:"seat_type_id" label:"Default seat type" validate:"gt=0"`
}

type AdminSeatRowTypesForm struct {
	RowLabels   []string `json:"row_label" form:"row_label" label:"Row" validate:"required,max=702,dive,alpha,uppercase,min=1,max=2"`
	SeatTypeIDs []int64  `json:"seat_type_id" form:"seat_type_id" label:"Seat type" validate:"required,max=702,dive,gt=0"`
}

type AdminSeatStatusResponse struct {
	IsActive bool `json:"is_active"`
}
