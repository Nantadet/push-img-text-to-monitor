package item

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const maxUploadImageSize int64 = 8 << 20

var errInvalidImageUpload = errors.New("invalid image upload")

var allowedUploadImageExts = map[string]bool{
	".gif":  true,
	".jpeg": true,
	".jpg":  true,
	".png":  true,
	".webp": true,
}

type Handler struct {
	svc *Service
	v   *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, v: validator.New()}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/items/preview", h.Preview)
	app.Get("/items/image", h.ImageProxy)
	app.Post("/items", h.Create)
	app.Get("/uploads/items/:file", h.UploadedImage)

	app.Get("/admin/items", h.List)
	app.Get("/admin/items/current", h.Current)
	app.Post("/admin/items/:id/show", h.Show)
	app.Post("/admin/items/current/skip", h.SkipCurrent)
	app.Post("/admin/items/current/add-time", h.AddTime)
	app.Post("/admin/items/current/reduce-time", h.ReduceTime)
	app.Post("/admin/items/current/finish", h.FinishCurrent)
	app.Delete("/admin/items/:id", h.Delete)

	app.Get("/display/current", h.Current)
	app.Get("/display/queue", h.Queue)
}

func (h *Handler) Preview(c fiber.Ctx) error {
	var dto PreviewDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.v.Struct(dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	preview, err := h.svc.Preview(c.Context(), dto.IGURL)
	if err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(preview)
}

func (h *Handler) ImageProxy(c fiber.Ctx) error {
	raw := strings.TrimSpace(c.Query("src"))
	if raw == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing image src"})
	}

	imageURL, err := neturl.Parse(raw)
	if err != nil || !isAllowedInstagramImageURL(imageURL) {
		return c.Status(400).JSON(fiber.Map{"error": "invalid image src"})
	}

	req, err := http.NewRequestWithContext(c.Context(), http.MethodGet, imageURL.String(), nil)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid image src"})
	}
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.instagram.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0 Safari/537.36")

	res, err := h.svc.preview.http.Do(req)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "unable to fetch instagram image"})
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return c.Status(502).JSON(fiber.Map{"error": "instagram image unavailable"})
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": "unable to read instagram image"})
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=300")
	if length := res.Header.Get("Content-Length"); length != "" {
		c.Set("Content-Length", length)
	}

	return c.Send(body)
}

func (h *Handler) Create(c fiber.Ctx) error {
	if c.IsMultipart() {
		return h.CreateMultipart(c)
	}

	var dto CreateItemDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.v.Struct(dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	item, err := h.svc.Create(c.Context(), dto)
	if err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(item)
}

func (h *Handler) CreateMultipart(c fiber.Ctx) error {
	sourceType := strings.TrimSpace(c.FormValue("sourceType"))
	if sourceType == "" {
		sourceType = SourceImage
	}
	if sourceType != SourceImage {
		return c.Status(400).JSON(fiber.Map{"error": "multipart submissions must use image sourceType"})
	}

	message := strings.TrimSpace(c.FormValue("message"))
	if message == "" {
		return c.Status(400).JSON(fiber.Map{"error": "message is required"})
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "image is required"})
	}

	imageURL, err := h.saveUploadedImage(c, file)
	if err != nil {
		status := 500
		if errors.Is(err, errInvalidImageUpload) {
			status = 400
		}
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}

	dto := CreateItemDTO{
		SourceType: SourceImage,
		IGImageURL: imageURL,
		Message:    message,
	}
	if err := h.v.Struct(dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	item, err := h.svc.Create(c.Context(), dto)
	if err != nil {
		return c.Status(422).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(item)
}

func (h *Handler) UploadedImage(c fiber.Ctx) error {
	fileName := filepath.Base(c.Params("file"))
	if fileName == "." || fileName == string(filepath.Separator) || fileName != c.Params("file") {
		return c.Status(400).JSON(fiber.Map{"error": "invalid image path"})
	}
	if !allowedUploadImageExts[strings.ToLower(filepath.Ext(fileName))] {
		return c.Status(404).JSON(fiber.Map{"error": "image not found"})
	}

	path := filepath.Join("uploads", "items", fileName)
	if _, err := os.Stat(path); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "image not found"})
	}

	c.Set("Cache-Control", "public, max-age=86400")
	return c.SendFile(path)
}

func (h *Handler) saveUploadedImage(c fiber.Ctx, file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", fmt.Errorf("%w: image is required", errInvalidImageUpload)
	}
	if file.Size <= 0 {
		return "", fmt.Errorf("%w: image is empty", errInvalidImageUpload)
	}
	if file.Size > maxUploadImageSize {
		return "", fmt.Errorf("%w: image must be 8MB or smaller", errInvalidImageUpload)
	}

	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	if contentType != "" && !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("%w: file must be an image", errInvalidImageUpload)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedUploadImageExts[ext] {
		return "", fmt.Errorf("%w: image must be jpg, png, webp, or gif", errInvalidImageUpload)
	}

	dir := filepath.Join("uploads", "items")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%d%s", time.Now().UTC().UnixNano(), ext)
	path := filepath.Join(dir, fileName)
	if err := c.SaveFile(file, path); err != nil {
		return "", err
	}

	return c.BaseURL() + "/uploads/items/" + fileName, nil
}

func (h *Handler) List(c fiber.Ctx) error {
	items, err := h.svc.List(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *Handler) Queue(c fiber.Ctx) error {
	items, err := h.svc.Queue(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(items)
}

func (h *Handler) Current(c fiber.Ctx) error {
	item, err := h.svc.Current(c.Context())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.JSON(nil)
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

func (h *Handler) Show(c fiber.Ctx) error {
	id, err := bson.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	item, err := h.svc.Show(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

func (h *Handler) SkipCurrent(c fiber.Ctx) error {
	item, err := h.svc.SkipCurrent(c.Context())
	if err != nil {
		if errors.Is(err, errNoCurrentItem) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

func (h *Handler) AddTime(c fiber.Ctx) error {
	var dto AdjustTimeDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.v.Struct(dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	item, err := h.svc.AddTime(c.Context(), dto.Minutes)
	if err != nil {
		if errors.Is(err, errNoCurrentItem) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

func (h *Handler) ReduceTime(c fiber.Ctx) error {
	var dto AdjustTimeDTO
	if err := c.Bind().Body(&dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := h.v.Struct(dto); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	item, err := h.svc.ReduceTime(c.Context(), dto.Minutes)
	if err != nil {
		if errors.Is(err, errNoCurrentItem) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

func (h *Handler) FinishCurrent(c fiber.Ctx) error {
	item, err := h.svc.FinishCurrent(c.Context())
	if err != nil {
		if errors.Is(err, errNoCurrentItem) {
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(item)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := bson.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}
