package push

import "mime/multipart"

type Push struct {
	Text string                `form:"text" json:"text"`
	File *multipart.FileHeader `form:"file" json:"-"`
}
