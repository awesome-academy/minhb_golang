package dto

type AdminTheaterForm struct {
	Name        string `json:"name" form:"name" label:"Name" validate:"required,max=200"`
	Address     string `json:"address" form:"address" label:"Address" validate:"required,max=500"`
	City        string `json:"city" form:"city" label:"City" validate:"required,max=100"`
	Phone       string `json:"phone" form:"phone" label:"Phone" validate:"max=30"`
	Description string `json:"description" form:"description" label:"Description" validate:"max=5000"`
	ImageURL    string `json:"image_url" form:"image_url" label:"Image URL" validate:"omitempty,url,max=2048"`
	IsActive    bool   `json:"is_active" form:"is_active"`
	UpdatedAt   string `json:"updated_at" form:"updated_at"`
}
