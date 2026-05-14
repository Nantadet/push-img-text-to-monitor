package item

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/HLLC-MFU/hllc-workshop-backend/ws"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var errNoCurrentItem = errors.New("no current displaying item")
var errMissingMessage = errors.New("message is required")

type Service struct {
	repo    *Repository
	preview *previewClient
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:    repo,
		preview: newPreviewClient(),
	}
}

func (s *Service) Preview(ctx context.Context, rawURL string) (*PreviewResponse, error) {
	return s.preview.Preview(ctx, rawURL)
}

func (s *Service) Create(ctx context.Context, dto CreateItemDTO) (*ItemResponse, error) {
	message := strings.TrimSpace(dto.Message)
	if message == "" {
		return nil, errMissingMessage
	}

	now := time.Now().UTC()
	it := &Item{
		SourceType:     normalizeSourceType(dto),
		Message:        message,
		Status:         StatusQueued,
		DisplayMinutes: 1,
		CreatedAt:      now,
	}

	switch it.SourceType {
	case SourceInstagram, SourceTiktok, SourceYoutube:
		url := strings.TrimSpace(dto.URL)
		if url == "" {
			url = strings.TrimSpace(dto.IGURL)
		}
		if url == "" {
			return nil, errors.New("url is required")
		}
		preview, err := s.Preview(ctx, url)
		if err != nil {
			return nil, err
		}
		it.IGURL = preview.IGURL
		it.IGImageURL = preview.IGImageURL
		it.IGUsername = preview.IGUsername
		it.VideoURL = preview.VideoURL
		it.AudioURL = preview.AudioURL
		it.EmbedURL = preview.EmbedURL
	case SourceImage:
		it.IGImageURL = strings.TrimSpace(dto.IGImageURL)
		it.IGUsername = strings.TrimSpace(dto.IGUsername)
		if it.IGImageURL == "" {
			return nil, errors.New("image is required")
		}
	case SourceText:
		// Text-only submissions intentionally carry no link or image.
	default:
		return nil, errors.New("invalid source type")
	}

	return s.insertAndBroadcast(ctx, it)
}

func (s *Service) insertAndBroadcast(ctx context.Context, it *Item) (*ItemResponse, error) {
	it, err := s.repo.Insert(ctx, it)
	if err != nil {
		return nil, err
	}

	ws.Broadcast("item-created", fiberMap("id", it.ID.Hex()))
	if _, err := s.repo.FindCurrent(ctx); errors.Is(err, mongo.ErrNoDocuments) {
		if _, err := s.promoteNextQueued(ctx); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
	}
	ws.Broadcast("queue-updated", nil)

	return toResponse(it), nil
}

func normalizeSourceType(dto CreateItemDTO) string {
	sourceType := strings.TrimSpace(dto.SourceType)
	if sourceType != "" {
		return sourceType
	}
	url := strings.TrimSpace(dto.URL)
	if url == "" {
		url = strings.TrimSpace(dto.IGURL)
	}
	if url != "" {
		platform := detectPlatform(url)
		switch platform {
		case platformInstagram:
			return SourceInstagram
		case platformTiktok:
			return SourceTiktok
		case platformYoutube:
			return SourceYoutube
		}
	}
	if strings.TrimSpace(dto.IGImageURL) != "" {
		return SourceImage
	}
	return SourceText
}

func (s *Service) List(ctx context.Context) ([]ItemResponse, error) {
	items, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return toResponses(items), nil
}

func (s *Service) Queue(ctx context.Context) ([]ItemResponse, error) {
	items, err := s.repo.FindQueue(ctx)
	if err != nil {
		return nil, err
	}
	return toResponses(items), nil
}

func (s *Service) Current(ctx context.Context) (*ItemResponse, error) {
	it, err := s.repo.FindCurrent(ctx)
	if err != nil {
		return nil, err
	}
	return toResponse(it), nil
}

func (s *Service) Show(ctx context.Context, id bson.ObjectID) (*ItemResponse, error) {
	now := time.Now().UTC()

	current, err := s.repo.FindCurrent(ctx)
	if err == nil && current.ID != id {
		if err := s.repo.UpdateByID(ctx, current.ID, bson.M{
			"status":     StatusDisplayed,
			"finishedAt": now,
		}); err != nil {
			return nil, err
		}
	}
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	it, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateByID(ctx, it.ID, bson.M{
		"status":      StatusDisplaying,
		"displayedAt": now,
		"finishedAt":  nil,
	}); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	ws.Broadcast("item-showing", fiberMap("id", updated.ID.Hex()))
	ws.Broadcast("queue-updated", nil)
	return toResponse(updated), nil
}

func (s *Service) SkipCurrent(ctx context.Context) (*ItemResponse, error) {
	next, err := s.finishCurrentWithStatus(ctx, StatusSkipped)
	if err != nil {
		return nil, err
	}
	ws.Broadcast("item-skipped", nil)
	return next, nil
}

func (s *Service) FinishCurrent(ctx context.Context) (*ItemResponse, error) {
	next, err := s.finishCurrentWithStatus(ctx, StatusDisplayed)
	if err != nil {
		return nil, err
	}
	ws.Broadcast("display-ended", nil)
	return next, nil
}

func (s *Service) AddTime(ctx context.Context, minutes int) (*ItemResponse, error) {
	return s.adjustCurrentTime(ctx, minutes, "item-time-added")
}

func (s *Service) ReduceTime(ctx context.Context, minutes int) (*ItemResponse, error) {
	return s.adjustCurrentTime(ctx, -minutes, "item-time-reduced")
}

func (s *Service) Delete(ctx context.Context, id bson.ObjectID) error {
	current, err := s.repo.FindCurrent(ctx)
	if err == nil && current.ID == id {
		if err := s.repo.DeleteByID(ctx, id); err != nil {
			return err
		}
		if _, err := s.promoteNextQueued(ctx); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
		ws.Broadcast("item-deleted", fiberMap("id", id.Hex()))
		ws.Broadcast("queue-updated", nil)
		return nil
	}
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return err
	}
	ws.Broadcast("item-deleted", fiberMap("id", id.Hex()))
	ws.Broadcast("queue-updated", nil)
	return nil
}

func (s *Service) finishCurrentWithStatus(ctx context.Context, status string) (*ItemResponse, error) {
	current, err := s.repo.FindCurrent(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errNoCurrentItem
		}
		return nil, err
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateByID(ctx, current.ID, bson.M{
		"status":     status,
		"finishedAt": now,
	}); err != nil {
		return nil, err
	}

	next, err := s.promoteNextQueued(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ws.Broadcast("queue-updated", nil)
			return nil, nil
		}
		return nil, err
	}

	ws.Broadcast("queue-updated", nil)
	return toResponse(next), nil
}

func (s *Service) promoteNextQueued(ctx context.Context) (*Item, error) {
	next, err := s.repo.FindOldestQueued(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateByID(ctx, next.ID, bson.M{
		"status":      StatusDisplaying,
		"displayedAt": now,
		"finishedAt":  nil,
	}); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, next.ID)
	if err != nil {
		return nil, err
	}

	ws.Broadcast("item-showing", fiberMap("id", updated.ID.Hex()))
	return updated, nil
}

func (s *Service) adjustCurrentTime(ctx context.Context, delta int, eventName string) (*ItemResponse, error) {
	current, err := s.repo.FindCurrent(ctx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errNoCurrentItem
		}
		return nil, err
	}

	newValue := current.displayMinutesValue() + delta
	if newValue < 1 {
		newValue = 1
	}

	if err := s.repo.UpdateByID(ctx, current.ID, bson.M{
		"displayMinutes": newValue,
		"displaySeconds": nil,
	}); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, current.ID)
	if err != nil {
		return nil, err
	}
	ws.Broadcast(eventName, fiberMap("id", updated.ID.Hex(), "displayMinutes", updated.displayMinutesValue()))
	return toResponse(updated), nil
}

func fiberMap(kv ...any) map[string]any {
	out := map[string]any{}
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			continue
		}
		out[key] = kv[i+1]
	}
	return out
}
