package item

type PreviewDTO struct {
	IGURL string `json:"igUrl" validate:"required,url,max=500"`
}

type CreateItemDTO struct {
	SourceType string `json:"sourceType" validate:"omitempty,oneof=instagram image text"`
	IGURL      string `json:"igUrl" validate:"omitempty,url,max=500"`
	IGImageURL string `json:"igImageUrl" validate:"omitempty,max=500"`
	IGUsername string `json:"igUsername" validate:"omitempty,max=120"`
	Message    string `json:"message" validate:"required,min=1,max=200"`
}

type AdjustTimeDTO struct {
	Minutes int `json:"minutes" validate:"required,min=1,max=120"`
}

type PreviewResponse struct {
	IGURL      string `json:"igUrl"`
	IGImageURL string `json:"igImageUrl"`
	IGUsername string `json:"igUsername"`
}
