package dto

type AdminCounterSaleForm struct {
	SeatIDs []int64 `json:"seat_id" form:"seat_id" label:"Seats" validate:"required,min=1,unique,dive,gt=0"`
}

type AdminConfirmBookingForm struct {
	Code       string `json:"code" form:"code" label:"Booking code" validate:"required,len=8"`
	ShowtimeID int64  `json:"showtime_id" form:"showtime_id"`
}
