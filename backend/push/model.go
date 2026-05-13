package push

import "mime/multipart"

type Push struct {
	ID      string                `json:"id" bson:"_id,omitempty"`
	Message string                `json:"message" bson:"message"`
	File    *multipart.FileHeader `form:"file" json:"-"`
}
