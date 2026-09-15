package dto

type AdminRoomForm struct {
	Name      string `json:"name" form:"name" label:"Name" validate:"required,max=100"`
	IsActive  bool   `json:"is_active" form:"is_active"`
	UpdatedAt string `json:"updated_at" form:"updated_at"`
}
