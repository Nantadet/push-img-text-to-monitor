package item

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	StatusQueued     = "queued"
	StatusDisplaying = "displaying"
	StatusSkipped    = "skipped"
	StatusDisplayed  = "displayed"

	SourceInstagram = "instagram"
	SourceImage     = "image"
	SourceText      = "text"
)

type Item struct {
	ID                   bson.ObjectID `bson:"_id,omitempty"`
	SourceType           string        `bson:"sourceType"`
	IGURL                string        `bson:"igUrl"`
	IGImageURL           string        `bson:"igImageUrl"`
	IGUsername           string        `bson:"igUsername"`
	Message              string        `bson:"message"`
	Status               string        `bson:"status"`
	DisplayMinutes       int           `bson:"displayMinutes,omitempty"`
	LegacyDisplaySeconds int           `bson:"displaySeconds,omitempty"`
	CreatedAt            time.Time     `bson:"createdAt"`
	DisplayedAt          *time.Time    `bson:"displayedAt,omitempty"`
	FinishedAt           *time.Time    `bson:"finishedAt,omitempty"`
}

type ItemResponse struct {
	ID             string     `json:"id"`
	SourceType     string     `json:"sourceType"`
	IGURL          string     `json:"igUrl"`
	IGImageURL     string     `json:"igImageUrl"`
	IGUsername     string     `json:"igUsername"`
	Message        string     `json:"message"`
	Status         string     `json:"status"`
	DisplayMinutes int        `json:"displayMinutes"`
	CreatedAt      time.Time  `json:"createdAt"`
	DisplayedAt    *time.Time `json:"displayedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
}

func (it *Item) displayMinutesValue() int {
	if it == nil {
		return 0
	}
	if it.DisplayMinutes > 0 {
		return it.DisplayMinutes
	}
	if it.LegacyDisplaySeconds > 0 {
		minutes := it.LegacyDisplaySeconds / 60
		if it.LegacyDisplaySeconds%60 != 0 {
			minutes++
		}
		if minutes < 1 {
			minutes = 1
		}
		return minutes
	}
	return 1
}

func (it *Item) sourceTypeValue() string {
	if it == nil {
		return ""
	}
	if it.SourceType != "" {
		return it.SourceType
	}
	if it.IGURL != "" {
		return SourceInstagram
	}
	if it.IGImageURL != "" {
		return SourceImage
	}
	return SourceText
}

func toResponse(it *Item) *ItemResponse {
	if it == nil {
		return nil
	}

	return &ItemResponse{
		ID:             it.ID.Hex(),
		SourceType:     it.sourceTypeValue(),
		IGURL:          it.IGURL,
		IGImageURL:     it.IGImageURL,
		IGUsername:     it.IGUsername,
		Message:        it.Message,
		Status:         it.Status,
		DisplayMinutes: it.displayMinutesValue(),
		CreatedAt:      it.CreatedAt,
		DisplayedAt:    it.DisplayedAt,
		FinishedAt:     it.FinishedAt,
	}
}

func toResponses(items []Item) []ItemResponse {
	out := make([]ItemResponse, 0, len(items))
	for _, it := range items {
		out = append(out, *toResponse(&it))
	}
	return out
}
