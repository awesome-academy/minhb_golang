package dto

import "cinema-booking/internal/models"

type TheaterListQuery struct {
	City string `json:"city" query:"city" validate:"omitempty,max=100"`
}

type TheaterResponse struct {
	ID          int64   `json:"id" example:"1"`
	Slug        string  `json:"slug" example:"cgv-vincom-center-1a2b3c4d"`
	Name        string  `json:"name" example:"CGV Vincom Center"`
	Address     string  `json:"address" example:"72 Le Thanh Ton, District 1"`
	City        string  `json:"city" example:"Ho Chi Minh"`
	Phone       *string `json:"phone" example:"+84 28 3823 0123"`
	Description *string `json:"description" example:"Multiplex with 8 screens in the city centre."`
	ImageURL    *string `json:"imageUrl" example:"https://cdn.example.com/theaters/cgv-vincom.jpg"`
}

type TheaterListResponse struct {
	Items []TheaterResponse `json:"items"`
}

func NewTheaterResponse(theater *models.Theater) TheaterResponse {
	return TheaterResponse{
		ID:          theater.ID,
		Slug:        theater.Slug,
		Name:        theater.Name,
		Address:     theater.Address,
		City:        theater.City,
		Phone:       theater.Phone,
		Description: theater.Description,
		ImageURL:    theater.ImageURL,
	}
}

func NewTheaterListResponse(theaters []models.Theater) TheaterListResponse {
	items := make([]TheaterResponse, 0, len(theaters))
	for i := range theaters {
		items = append(items, NewTheaterResponse(&theaters[i]))
	}
	return TheaterListResponse{Items: items}
}
