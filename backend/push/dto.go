package push

import "mime/multipart"

type CreatePushDTO struct {
	Message string                `json:"message" validate:"required,min=5"`
	File    *multipart.FileHeader `form:"file" json:"-" validate:"required"`
}

type UpdatePushDTO struct {
	Message *string               `json:"message" validate:"omitempty,min=5"`
	File    *multipart.FileHeader `form:"file" json:"-" validate:"required"`
}
