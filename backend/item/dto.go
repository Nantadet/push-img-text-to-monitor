package item

type PreviewDTO struct {
	URL   string `json:"url" validate:"omitempty,url,max=500"`
	IGURL string `json:"igUrl" validate:"omitempty,url,max=500"`
}

func (d PreviewDTO) effectiveURL() string {
	if d.URL != "" {
		return d.URL
	}
	return d.IGURL
}

type CreateItemDTO struct {
	SourceType string `json:"sourceType" validate:"omitempty,oneof=instagram tiktok youtube image text"`
	URL        string `json:"url" validate:"omitempty,url,max=500"`
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
	VideoURL   string `json:"videoUrl"`
	AudioURL   string `json:"audioUrl"`
	EmbedURL   string `json:"embedUrl"`
}
